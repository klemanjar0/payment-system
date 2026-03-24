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
	CreateRefreshToken(ctx context.Context, token *RefreshToken) (*RefreshToken, error)
	GetRefreshToken(ctx context.Context, tokenID string) (*RefreshToken, error)
	// ConsumeRefreshToken atomically marks a token as used (sets last_used_at).
	// Returns ErrInvalidRefreshToken if the token was already consumed, revoked, or expired.
	ConsumeRefreshToken(ctx context.Context, tokenID string) (*RefreshToken, error)
	RevokeTokenFamily(ctx context.Context, tokenID string) error
	RevokeAllUserTokens(ctx context.Context, userID string) error
}
