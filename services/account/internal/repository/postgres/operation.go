package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	utilid "github.com/klemanjar0/payment-system/pkg/util_id"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	"github.com/klemanjar0/payment-system/services/account/internal/repository/postgres/sqlc"
)


type OperationRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewOperationRepository(pool *pgxpool.Pool) *OperationRepository {
	return &OperationRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *OperationRepository) Create(ctx context.Context, op *domain.Operation) error {
	entity := op.ToSQLC()

	_, err := r.queries.CreateOperation(ctx, sqlc.CreateOperationParams{
		AccountID:     entity.AccountID,
		Type:          entity.Type,
		Amount:        entity.Amount,
		BalanceAfter:  entity.BalanceAfter,
		TransactionID: entity.TransactionID,
		Description:   entity.Description,
		CreatedAt:     entity.CreatedAt,
	})
	return err
}

func (r *OperationRepository) GetByAccountID(
	ctx context.Context,
	accountID string,
	limit, offset int,
) ([]*domain.Operation, int, error) {
	pgAccountID := utilid.FromString(accountID).AsPgUUID()

	total, err := r.queries.CountOperationsByAccountID(ctx, pgAccountID)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.GetOperationsByAccountID(ctx, sqlc.GetOperationsByAccountIDParams{
		AccountID: pgAccountID,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Operation, len(rows))
	for i, row := range rows {
		result[i] = domain.OperationFromSQLC(&row)
	}

	return result, int(total), nil
}
