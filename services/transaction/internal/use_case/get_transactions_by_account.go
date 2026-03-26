package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
)

type GetTransactionsByAccountUseCase struct {
	repo domain.TransactionRepository
}

func NewGetTransactionsByAccountUseCase(repo domain.TransactionRepository) *GetTransactionsByAccountUseCase {
	return &GetTransactionsByAccountUseCase{repo: repo}
}

func (uc *GetTransactionsByAccountUseCase) Execute(ctx context.Context, accountID string, limit, offset int) ([]*domain.Transaction, int, error) {
	return uc.repo.GetByAccountID(ctx, accountID, limit, offset)
}
