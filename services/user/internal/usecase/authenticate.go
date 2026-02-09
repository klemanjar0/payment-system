package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/services/user/internal/domain"
)

type AuthenticateUseCase struct {
	repo     domain.UserRepository
	tokenSvc TokenService
}

func NewAuthenticateUseCase(
	repo domain.UserRepository,
	tokenSvc TokenService,
) *AuthenticateUseCase {
	return &AuthenticateUseCase{
		repo:     repo,
		tokenSvc: tokenSvc,
	}
}

type AuthenticateInput struct {
	Email    string
	Password string
}

type AuthenticateOutput struct {
	UserID       string
	AccessToken  string
	RefreshToken string
}

func (uc *AuthenticateUseCase) Execute(ctx context.Context, input AuthenticateInput) (*AuthenticateOutput, error) {
	user, err := uc.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.CheckPassword(input.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	if user.Status == domain.UserStatusBlocked {
		return nil, domain.ErrUserBlocked
	}

	if user.Status != domain.UserStatusActive {
		return nil, domain.ErrUserNotActive
	}

	accessToken, err := uc.tokenSvc.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	refreshToken, err := uc.tokenSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, domain.ErrInternal
	}

	return &AuthenticateOutput{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
