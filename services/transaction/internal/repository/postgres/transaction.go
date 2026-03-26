package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/klemanjar0/payment-system/pkg/logger"
	utilid "github.com/klemanjar0/payment-system/pkg/util_id"
	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
	"github.com/klemanjar0/payment-system/services/transaction/internal/repository/postgres/sqlc"
)

type TransactionRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	row, err := r.queries.CreateTransaction(ctx, tx.ToSQLC())
	if err != nil {
		return err
	}
	tx.ID = utilid.FromPg(row.ID).AsString()
	tx.CreatedAt = row.CreatedAt.Time
	tx.UpdatedAt = row.UpdatedAt.Time
	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
	row, err := r.queries.GetTransactionByID(ctx, utilid.FromString(id).AsPgUUID())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTransactionNotFound
		}
		return nil, err
	}
	return domain.TransactionFromSQLC(row), nil
}

func (r *TransactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	row, err := r.queries.GetTransactionByIdempotencyKey(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTransactionNotFound
		}
		return nil, err
	}
	return domain.TransactionFromSQLC(row), nil
}

func (r *TransactionRepository) GetByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*domain.Transaction, int, error) {
	pgAccountID := utilid.FromString(accountID).AsPgUUID()

	count, err := r.queries.CountTransactionsByAccount(ctx, pgAccountID)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.GetTransactionsByAccount(ctx, sqlc.GetTransactionsByAccountParams{
		FromAccountID: pgAccountID,
		Limit:         int32(limit),
		Offset:        int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	txs := make([]*domain.Transaction, 0, len(rows))
	for _, row := range rows {
		txs = append(txs, domain.TransactionFromSQLC(row))
	}

	return txs, int(count), nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	_, err := r.queries.UpdateTransactionStatus(ctx, sqlc.UpdateTransactionStatusParams{
		ID:     utilid.FromString(id).AsPgUUID(),
		Status: sqlc.TransactionStatus(status),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrTransactionNotFound
		}
		logger.Error("UpdateStatus failed", "err", err, "id", id)
		return err
	}
	return nil
}

func (r *TransactionRepository) UpdateStatusWithReason(ctx context.Context, id string, status domain.TransactionStatus, reason string) error {
	_, err := r.queries.UpdateTransactionStatusWithReason(ctx, sqlc.UpdateTransactionStatusWithReasonParams{
		ID:            utilid.FromString(id).AsPgUUID(),
		Status:        sqlc.TransactionStatus(status),
		FailureReason: reason,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrTransactionNotFound
		}
		logger.Error("UpdateStatusWithReason failed", "err", err, "id", id)
		return err
	}
	return nil
}
