package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
)

// CommandPublisher publishes saga commands to the account service via Kafka.
type CommandPublisher interface {
	Publish(ctx context.Context, eventType string, payload any) error
}

// AccountInfo is a lightweight view of an account returned by the account service gRPC client.
type AccountInfo struct {
	ID       string
	UserID   string
	Currency string
	Status   string
}

// UserInfo is a lightweight view of a user returned by the user service gRPC client.
type UserInfo struct {
	ID     string
	Status string
}

// AccountServiceClient abstracts the gRPC call to account service.
type AccountServiceClient interface {
	GetAccount(ctx context.Context, accountID string) (*AccountInfo, error)
}

// UserServiceClient abstracts the gRPC call to user service.
type UserServiceClient interface {
	GetUser(ctx context.Context, userID string) (*UserInfo, error)
}

// Use case interfaces.

type TransactionCreator interface {
	Execute(ctx context.Context, params *CreateTransferParams) (*domain.Transaction, error)
}

type TransactionGetter interface {
	Execute(ctx context.Context, id string) (*domain.Transaction, error)
}

type TransactionsByAccountGetter interface {
	Execute(ctx context.Context, accountID string, limit, offset int) ([]*domain.Transaction, int, error)
}

// CreateTransferParams holds the input for creating a new transfer.
type CreateTransferParams struct {
	IdempotencyKey string
	FromAccountID  string
	ToAccountID    string
	Amount         int64
	Currency       string
	Description    string
}
