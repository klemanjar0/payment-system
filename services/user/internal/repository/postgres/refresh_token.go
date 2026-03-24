package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/klemanjar0/payment-system/pkg/logger"
	utilid "github.com/klemanjar0/payment-system/pkg/util_id"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
	"github.com/klemanjar0/payment-system/services/user/internal/repository/postgres/sqlc"
)

type RefreshTokenRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	queries := sqlc.New(pool)
	return &RefreshTokenRepository{pool: pool, queries: queries}
}

func (r *RefreshTokenRepository) CreateRefreshToken(
	ctx context.Context,
	payload *domain.RefreshToken,
) (*domain.RefreshToken, error) {
	entity, err := r.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:      utilid.FromString(payload.UserID).AsPgUUID(),
		DeviceInfo:  pgtype.Text{String: payload.DeviceInfo, Valid: payload.DeviceInfo != ""},
		ExpiresAt:   pgtype.Timestamptz{Time: payload.ExpiresAt, Valid: true},
		RotatedFrom: utilid.FromString(payload.RotatedFrom).AsPgUUID(),
	})
	if err != nil {
		logger.Error("failed to create refresh token", "err", err, "user_id", payload.UserID)
		return nil, err
	}

	logger.Info("refresh token created", "token_id", entity.ID, "user_id", payload.UserID)
	return domain.NewRefreshTokenOfSql(&entity)
}

func (r *RefreshTokenRepository) GetRefreshToken(ctx context.Context, tokenID string) (*domain.RefreshToken, error) {
	entity, err := r.queries.GetRefreshToken(ctx, utilid.FromString(tokenID).AsPgUUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidRefreshToken
		}
		logger.Error("failed to get refresh token", "err", err, "token_id", tokenID)
		return nil, domain.ErrInternal
	}
	return domain.NewRefreshTokenOfSql(&entity)
}

// ConsumeRefreshToken atomically claims a never-used token (last_used_at IS NULL).
// Returns ErrInvalidRefreshToken if the token was already consumed, revoked, or expired.
func (r *RefreshTokenRepository) ConsumeRefreshToken(ctx context.Context, tokenID string) (*domain.RefreshToken, error) {
	entity, err := r.queries.ConsumeRefreshToken(ctx, utilid.FromString(tokenID).AsPgUUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Token was already consumed, revoked, expired, or doesn't exist.
			return nil, domain.ErrInvalidRefreshToken
		}
		logger.Error("failed to consume refresh token", "err", err, "token_id", tokenID)
		return nil, domain.ErrInternal
	}
	logger.Info("refresh token consumed", "token_id", tokenID)
	return domain.NewRefreshTokenOfSql(&entity)
}

// RevokeTokenFamily revokes the given token and all tokens descended from it.
func (r *RefreshTokenRepository) RevokeTokenFamily(ctx context.Context, tokenID string) error {
	err := r.queries.RevokeTokenFamily(ctx, utilid.FromString(tokenID).AsPgUUID())
	if err != nil {
		logger.Error("failed to revoke token family", "err", err, "token_id", tokenID)
		return err
	}
	logger.Info("token family revoked", "root_token_id", tokenID)
	return nil
}

func (r *RefreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	err := r.queries.RevokeAllUserTokens(ctx, utilid.FromString(userID).AsPgUUID())
	if err != nil {
		logger.Error("failed to revoke all user tokens", "err", err, "user_id", userID)
		return err
	}
	logger.Info("all user refresh tokens revoked", "user_id", userID)
	return nil
}
