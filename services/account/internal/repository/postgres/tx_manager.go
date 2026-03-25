package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	"github.com/klemanjar0/payment-system/services/account/internal/repository/postgres/sqlc"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func (m *TxManager) ExecTx(ctx context.Context, fn func(domain.TxRepositories) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := sqlc.New(tx)
	repos := domain.TxRepositories{
		Accounts: &AccountRepository{queries: q},
		//Holds:      &HoldRepository{queries: q},
		//Operations: &TxOperationRepository{queries: q},
	}
	if err := fn(repos); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
