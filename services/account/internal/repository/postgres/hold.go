package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	utilid "github.com/klemanjar0/payment-system/pkg/util_id"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	"github.com/klemanjar0/payment-system/services/account/internal/repository/postgres/sqlc"
)

type HoldRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewHoldRepository(pool *pgxpool.Pool) *HoldRepository {
	return &HoldRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *HoldRepository) Create(ctx context.Context, hold *domain.Hold) error {
	entity := hold.ToSQLC()
	_, err := r.queries.CreateHold(ctx, sqlc.CreateHoldParams{
		AccountID:     entity.AccountID,
		TransactionID: entity.TransactionID,
		Amount:        entity.Amount,
		Description:   entity.Description,
		Status:        sqlc.HoldStatusActive,
		CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *HoldRepository) GetByID(ctx context.Context, id string) (*domain.Hold, error) {
	hold, err := r.queries.GetHoldByID(ctx, utilid.FromString(id).AsPgUUID())

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrHoldNotFound
	}

	if err != nil {
		return nil, err
	}

	return domain.HoldFromSQLC(&hold), nil
}

func (r *HoldRepository) GetByTransactionID(
	ctx context.Context,
	accountId, transactionId string,
) (*domain.Hold, error) {
	hold, err := r.queries.GetHoldByTransactionID(ctx, sqlc.GetHoldByTransactionIDParams{
		AccountID:     utilid.FromString(accountId).AsPgUUID(),
		TransactionID: utilid.FromString(transactionId).AsPgUUID(),
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrHoldNotFound
	}

	if err != nil {
		return nil, err
	}

	return domain.HoldFromSQLC(&hold), nil
}

func (r *HoldRepository) GetByTransactionIDForUpdate(
	ctx context.Context,
	accountId, transactionId string,
) (*domain.Hold, error) {
	hold, err := r.queries.GetHoldByTransactionIDForUpdate(ctx,
		sqlc.GetHoldByTransactionIDForUpdateParams{
			AccountID:     utilid.FromString(accountId).AsPgUUID(),
			TransactionID: utilid.FromString(transactionId).AsPgUUID(),
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrHoldNotFound
	}

	if err != nil {
		return nil, err
	}

	return domain.HoldFromSQLC(&hold), nil
}

func (r *HoldRepository) GetActiveByAccountID(
	ctx context.Context,
	accountId string,
) ([]*domain.Hold, error) {
	holds, err := r.queries.GetActiveHoldsByAccountID(
		ctx,
		utilid.FromString(accountId).AsPgUUID(),
	)

	if err != nil {
		return nil, err
	}

	result := make([]*domain.Hold, len(holds))
	for i, row := range holds {
		result[i] = domain.HoldFromSQLC(&row)
	}

	return result, nil
}

func (r *HoldRepository) Update(ctx context.Context, hold *domain.Hold) error {
	entity := hold.ToSQLC()

	_, err := r.queries.UpdateHold(ctx, sqlc.UpdateHoldParams{
		ID:         entity.ID,
		Status:     entity.Status,
		ReleasedAt: entity.ReleasedAt,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrHoldNotFound
	}

	return err
}
