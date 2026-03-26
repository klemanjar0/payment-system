package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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

func (r *OperationRepository) GetByTransactionIDAndType(
	ctx context.Context,
	accountID, transactionID string,
	opType domain.OperationType,
) (*domain.Operation, error) {
	row, err := r.queries.GetOperationByTransactionIDAndType(ctx, sqlc.GetOperationByTransactionIDAndTypeParams{
		AccountID:     utilid.FromString(accountID).AsPgUUID(),
		TransactionID: utilid.FromString(transactionID).AsPgUUID(),
		Type:          sqlc.OperationType(opType),
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOperationNotFound
	}

	if err != nil {
		return nil, err
	}

	return domain.OperationFromSQLC(&row), nil
}
