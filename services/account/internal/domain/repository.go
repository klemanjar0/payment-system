package domain

import (
	"context"
)

type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByUserID(ctx context.Context, userID string) ([]*Account, error)
	GetByUserAndCurrency(ctx context.Context, userID string, currency Currency) (*Account, error)
	Update(ctx context.Context, account *Account) error

	// called inside ExecTx callback — tx is embedded in the repo
	GetByIDForUpdate(ctx context.Context, id string) (*Account, error)
}

type HoldRepository interface {
	Create(ctx context.Context, hold *Hold) error
	GetByID(ctx context.Context, id string) (*Hold, error)
	GetByTransactionID(ctx context.Context, accountID, transactionID string) (*Hold, error)
	GetActiveByAccountID(ctx context.Context, accountID string) ([]*Hold, error)
	Update(ctx context.Context, hold *Hold) error

	GetByTransactionIDForUpdate(ctx context.Context, accountID, transactionID string) (*Hold, error)
}

type OperationRepository interface {
	Create(ctx context.Context, op *Operation) error
	GetByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*Operation, int, error)
	GetByTransactionIDAndType(ctx context.Context, accountID, transactionID string, opType OperationType) (*Operation, error)
}

// TxRepositories groups all repositories scoped to a single transaction.
// Passed to the ExecTx callback — each repo shares the same underlying pgx.Tx.
type TxRepositories struct {
	Accounts   AccountRepository
	Holds      HoldRepository
	Operations OperationRepository
}

type TransactionManager interface {
	ExecTx(ctx context.Context, fn func(tx TxRepositories) error) error
}
