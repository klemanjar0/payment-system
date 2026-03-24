package usecase

import (
	"context"
	"time"

	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
)

// TokenBlacklist is satisfied by pkg/tokenblacklist.Blacklist.
type TokenBlacklist interface {
	Blacklist(ctx context.Context, jti string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}

type LogoutUseCase struct {
	tokenRepo domain.RefreshTokenRepository
	tokenSvc  auth.TokenService
	blacklist TokenBlacklist
	auditLog  UserAuditLogger
}

type LogoutInput struct {
	UserID       string // extracted from auth middleware
	AccessToken  string // raw JWT string to blacklist
	RefreshToken string // optional: when present, only this family is revoked
}

func NewLogoutUseCase(
	tokenRepo domain.RefreshTokenRepository,
	tokenSvc auth.TokenService,
	blacklist TokenBlacklist,
	auditLog UserAuditLogger,
) *LogoutUseCase {
	return &LogoutUseCase{
		tokenRepo: tokenRepo,
		tokenSvc:  tokenSvc,
		blacklist: blacklist,
		auditLog:  auditLog,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, input LogoutInput) error {
	// 1. Blacklist the current access token for its remaining lifetime.
	if input.AccessToken != "" {
		if claims, err := uc.tokenSvc.ValidateAccessToken(input.AccessToken); err == nil {
			if ttl := time.Until(claims.ExpiresAt.Time); ttl > 0 {
				_ = uc.blacklist.Blacklist(ctx, claims.ID, ttl)
			}
		}
	}

	// 2. Revoke refresh tokens.
	//    If a specific refresh token is provided, revoke only its family (targeted logout).
	//    Otherwise revoke every active token for the user (logout all sessions).
	if input.RefreshToken != "" {
		if claims, err := uc.tokenSvc.ValidateRefreshToken(input.RefreshToken); err == nil {
			_ = uc.tokenRepo.RevokeTokenFamily(ctx, claims.TokenID)
		} else {
			// Refresh token is already expired/invalid; fall back to revoking all.
			_ = uc.tokenRepo.RevokeAllUserTokens(ctx, input.UserID)
		}
	} else {
		_ = uc.tokenRepo.RevokeAllUserTokens(ctx, input.UserID)
	}

	uc.auditLog.LogTokenRevoked(ctx, input.UserID, "logout")
	return nil
}
