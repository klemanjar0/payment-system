package postgres

import (
	"context"

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

	if err == nil {
		logger.Info("RefreshTokenRepository->CreateRefreshToken. entity created", "refresh token", entity)
	} else {
		logger.Error("failed to create refresh token", "err", err, "user_id", payload.UserID)
		return nil, err
	}

	return domain.NewRefreshTokenOfSql(&entity)
}

func (r *RefreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userId, reason string) error {
	err := r.queries.RevokeAllUserTokens(ctx, utilid.FromString(userId).AsPgUUID())

	if err != nil {
		logger.Error("failed to revoke all user tokens in database", "err", err, "user_id", userId)
		return err
	}

	logger.Info("all user refresh tokens revoked in database", "user_id", userId, "reason", reason)

	return nil
}
