package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/services/account/internal/domain"
)

type GetAccountUseCase struct {
	accounts domain.AccountRepository
}

func NewGetAccountUseCase(accounts domain.AccountRepository) *GetAccountUseCase {
	return &GetAccountUseCase{accounts: accounts}
}

func (uc *GetAccountUseCase) Execute(ctx context.Context, id string) (*domain.Account, error) {
	return uc.accounts.GetByID(ctx, id)
}
