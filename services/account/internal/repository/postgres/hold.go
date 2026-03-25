package postgres

import (
	"github.com/klemanjar0/payment-system/services/account/internal/repository/postgres/sqlc"
)

type HoldRepository struct {
	queries *sqlc.Queries
}

func NewHoldRepository(queries *sqlc.Queries) *HoldRepository {
	return &HoldRepository{
		queries: queries,
	}
}
