package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/services/user/internal/domain"
)

type ChangePasswordUseCase struct {
	repo domain.UserRepository
}

func NewChangePasswordUseCase(repo domain.UserRepository) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{repo: repo}
}

type ChangePasswordInput struct {
	UserID      string
	OldPassword string
	NewPassword string
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, input ChangePasswordInput) error {
	user, err := uc.repo.GetByID(ctx, input.UserID)
	if err != nil {
		return domain.ErrUserNotFound
	}

	if err := user.ChangePassword(input.OldPassword, input.NewPassword); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, user); err != nil {
		return domain.ErrInternal
	}

	return nil
}
