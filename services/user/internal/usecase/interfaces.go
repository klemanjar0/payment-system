package usecase

import (
	"context"
	"time"
)

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload any) error
}

// UserAuditLogger defines the audit logging operations available to use cases.
// All methods are fire-and-forget: implementations must not block the caller.
type UserAuditLogger interface {
	LogUserCreated(ctx context.Context, userID, email string)
	LogLoginSuccess(ctx context.Context, userID, email string)
	LogLoginFailure(ctx context.Context, email, reason string)
	LogPasswordChanged(ctx context.Context, userID string)
	LogPasswordChangeFailed(ctx context.Context, userID, reason string)
	LogTokenRevoked(ctx context.Context, userID, reason string)
}

type UserCreatedEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserUpdatedEvent struct {
	UserID    string    `json:"user_id"`
	UpdatedAt time.Time `json:"updated_at"`
}
