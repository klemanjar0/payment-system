package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
)

type GetTransactionUseCase struct {
	repo domain.TransactionRepository
}

func NewGetTransactionUseCase(repo domain.TransactionRepository) *GetTransactionUseCase {
	return &GetTransactionUseCase{repo: repo}
}

func (uc *GetTransactionUseCase) Execute(ctx context.Context, id string) (*domain.Transaction, error) {
	return uc.repo.GetByID(ctx, id)
}
