package domain

import "context"

type TransactionRepository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id string) (*Transaction, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*Transaction, error)
	GetByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*Transaction, int, error)
	UpdateStatus(ctx context.Context, id string, status TransactionStatus) error
	UpdateStatusWithReason(ctx context.Context, id string, status TransactionStatus, reason string) error
}
