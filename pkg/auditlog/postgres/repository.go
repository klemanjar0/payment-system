package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/klemanjar0/payment-system/pkg/auditlog"
)

// Repository persists audit events to a PostgreSQL table.
type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, event auditlog.Event) error {
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("auditlog/postgres: marshal metadata: %w", err)
	}

	const q = `
		INSERT INTO audit_logs (id, service, action, actor_id, target_id, status, metadata, error, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = r.pool.Exec(ctx, q,
		event.ID, event.Service, event.Action,
		event.ActorID, event.TargetID, event.Status,
		metadata, event.Error, event.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("auditlog/postgres: insert failed: %w", err)
	}
	return nil
}

var _ auditlog.Repository = (*Repository)(nil)
