package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/services/account/internal/domain"
)

type GetAccountsByUserUseCase struct {
	accounts domain.AccountRepository
}

func NewGetAccountsByUserUseCase(accounts domain.AccountRepository) *GetAccountsByUserUseCase {
	return &GetAccountsByUserUseCase{accounts: accounts}
}

func (uc *GetAccountsByUserUseCase) Execute(ctx context.Context, userID string) ([]*domain.Account, error) {
	return uc.accounts.GetByUserID(ctx, userID)
}
