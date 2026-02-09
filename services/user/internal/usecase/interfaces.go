package usecase

import (
	"context"
	"time"
)

type TokenService interface {
	GenerateAccessToken(userID string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateAccessToken(token string) (userID string, err error)
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload interface{}) error
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
