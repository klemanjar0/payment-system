package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/services/user/internal/domain"
)

type GetUserUseCase struct {
	repo domain.UserRepository
}

func NewGetUserUseCase(repo domain.UserRepository) *GetUserUseCase {
	return &GetUserUseCase{repo: repo}
}

func (uc *GetUserUseCase) ExecuteByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

func (uc *GetUserUseCase) ExecuteByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}
