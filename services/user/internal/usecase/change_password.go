package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/services/user/internal/domain"
)

type ChangePasswordUseCase struct {
	repo     domain.UserRepository
	auditLog UserAuditLogger
}

func NewChangePasswordUseCase(
	repo domain.UserRepository,
	auditLog UserAuditLogger,
) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{repo: repo, auditLog: auditLog}
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
		uc.auditLog.LogPasswordChangeFailed(ctx, input.UserID, err.Error())
		return err
	}

	if err := uc.repo.Update(ctx, user); err != nil {
		uc.auditLog.LogPasswordChangeFailed(ctx, input.UserID, "database update failed")
		return domain.ErrInternal
	}

	uc.auditLog.LogPasswordChanged(ctx, input.UserID)
	return nil
}
