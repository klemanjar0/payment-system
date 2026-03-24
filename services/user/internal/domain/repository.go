package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	Update(ctx context.Context, user *User) error
	Deactivate(ctx context.Context, userID string) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type CachedUserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
}

type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, id string) (*RefreshToken, error)
	RevokeAllUserTokens(ctx context.Context, userId, reason string) error
}
