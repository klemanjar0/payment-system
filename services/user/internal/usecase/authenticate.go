package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/pkg/auth"
	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
)

type LoginParams struct {
	repo     domain.UserRepository
	tokenSvc auth.TokenService
	auditLog UserAuditLogger
}

type AuthenticateUseCase struct {
	*LoginParams
}

type AuthenticateInput struct {
	Email    string
	Password string
}

type AuthenticateResult struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

func NewAuthenticateUseCase(
	repo domain.UserRepository,
	tokenSvc auth.TokenService,
	auditLog UserAuditLogger,
) *AuthenticateUseCase {
	return &AuthenticateUseCase{
		LoginParams: &LoginParams{
			repo:     repo,
			tokenSvc: tokenSvc,
			auditLog: auditLog,
		},
	}
}

func (uc *AuthenticateUseCase) Execute(ctx context.Context, input AuthenticateInput) (*AuthenticateResult, error) {
	user, err := uc.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		uc.auditLog.LogLoginFailure(ctx, input.Email, domain.ErrInvalidCredentials.Error())
		return nil, domain.ErrInvalidCredentials
	}

	if !user.CheckPassword(input.Password) {
		uc.auditLog.LogLoginFailure(ctx, input.Email, domain.ErrInvalidCredentials.Error())
		return nil, domain.ErrInvalidCredentials
	}

	if user.Status == domain.UserStatusBlocked {
		uc.auditLog.LogLoginFailure(ctx, input.Email, domain.ErrUserBlocked.Error())
		return nil, domain.ErrUserBlocked
	}

	if user.Status != domain.UserStatusActive {
		uc.auditLog.LogLoginFailure(ctx, input.Email, domain.ErrUserNotActive.Error())
		return nil, domain.ErrUserNotActive
	}

	accessToken, err := uc.tokenSvc.GenerateAccessToken(user.ID)
	if err != nil {
		logger.Error("failed to generate access token", "err", err, "userId", user.ID)
		uc.auditLog.LogLoginFailure(ctx, input.Email, "token generation failed")
		return nil, domain.ErrInternal
	}

	refreshToken, err := uc.tokenSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		uc.auditLog.LogLoginFailure(ctx, input.Email, "refresh token generation failed")
		return nil, domain.ErrInternal
	}

	uc.auditLog.LogLoginSuccess(ctx, user.ID, user.Email)

	return &AuthenticateResult{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
