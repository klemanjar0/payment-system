package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/services/account/internal/domain"
)

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload any) error
}

// Read/write use cases used by gRPC delivery.

type AccountCreator interface {
	Execute(ctx context.Context, params *CreateAccountUseCaseParams) (*domain.Account, error)
}

type AccountGetter interface {
	Execute(ctx context.Context, id string) (*domain.Account, error)
}

type AccountsByUserGetter interface {
	Execute(ctx context.Context, userID string) ([]*domain.Account, error)
}

// Saga use cases used by Kafka consumer.

type HoldPlacer interface {
	Execute(ctx context.Context, params *PlaceHoldUseCaseParams) error
}

type HoldExecutor interface {
	Execute(ctx context.Context, params *ExecuteHoldUseCaseParams) error
}

type HoldReleaser interface {
	Execute(ctx context.Context, params *ReleaseHoldUseCaseParams) error
}

type AccountCreditor interface {
	Execute(ctx context.Context, params *CreditUseCaseParams) error
}
