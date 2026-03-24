package usecase

import (
	"context"
	"time"

	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
)

type RefreshTokenUseCase struct {
	tokenRepo domain.RefreshTokenRepository
	tokenSvc  auth.TokenService
	auditLog  UserAuditLogger
}

type RefreshTokenInput struct {
	RefreshToken string
	DeviceInfo   string
}

type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
}

func NewRefreshTokenUseCase(
	tokenRepo domain.RefreshTokenRepository,
	tokenSvc auth.TokenService,
	auditLog UserAuditLogger,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{tokenRepo: tokenRepo, tokenSvc: tokenSvc, auditLog: auditLog}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshTokenInput) (*RefreshTokenResult, error) {
	// 1. Validate the refresh JWT — claims.TokenID holds the DB token UUID.
	claims, err := uc.tokenSvc.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	oldTokenID := claims.TokenID

	// 2. Atomically consume the token (WHERE last_used_at IS NULL).
	//    If this fails the token was already used (reuse attack) or is otherwise invalid.
	consumed, err := uc.tokenRepo.ConsumeRefreshToken(ctx, oldTokenID)
	if err != nil {
		// Revoke the entire family — covers the reuse-attack case where last_used_at
		// was already set by a legitimate earlier rotation.
		_ = uc.tokenRepo.RevokeTokenFamily(ctx, oldTokenID)
		return nil, domain.ErrInvalidRefreshToken
	}

	// 3. Issue a new token row (rotated_from points to the consumed one).
	deviceInfo := input.DeviceInfo
	if deviceInfo == "" {
		deviceInfo = consumed.DeviceInfo
	}

	newDbToken, err := uc.tokenRepo.CreateRefreshToken(ctx, &domain.RefreshToken{
		UserID:      consumed.UserID,
		DeviceInfo:  deviceInfo,
		ExpiresAt:   time.Now().Add(auth.DefaultRefreshTokenExpiry),
		RotatedFrom: oldTokenID,
	})
	if err != nil {
		return nil, domain.ErrInternal
	}

	// 4. Sign JWTs — refresh JWT carries the new DB UUID, access JWT carries the user ID.
	newRefreshJWT, err := uc.tokenSvc.GenerateRefreshToken(newDbToken.TokenID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	newAccessToken, err := uc.tokenSvc.GenerateAccessToken(consumed.UserID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return &RefreshTokenResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshJWT,
	}, nil
}
