package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/klemanjar0/payment-system/pkg/money"
	utilid "github.com/klemanjar0/payment-system/pkg/util_id"
	"github.com/klemanjar0/payment-system/services/account/internal/repository/postgres/sqlc"
)

type HoldStatus string

const (
	HoldStatusActive   HoldStatus = "active"
	HoldStatusReleased HoldStatus = "released"
	HoldStatusExecuted HoldStatus = "executed" // fired from hold
)

type Hold struct {
	ID            string
	AccountID     string
	TransactionID string // idempotency key
	Amount        money.Amount
	Description   string
	Status        HoldStatus
	CreatedAt     time.Time
	ReleasedAt    *time.Time
}

func NewHold(accountID, transactionID string, amount money.Amount, description string) *Hold {
	return &Hold{
		AccountID:     accountID,
		TransactionID: transactionID,
		Amount:        amount,
		Description:   description,
		Status:        HoldStatusActive,
		CreatedAt:     time.Now(),
	}
}

func (h *Hold) IsActive() bool {
	return h.Status == HoldStatusActive
}

func (h *Hold) Release() {
	now := time.Now()
	h.Status = HoldStatusReleased
	h.ReleasedAt = &now
}

func (h *Hold) Execute() {
	now := time.Now()
	h.Status = HoldStatusExecuted
	h.ReleasedAt = &now
}

func HoldFromSQLC(entity *sqlc.Hold) *Hold {
	var releasedAt *time.Time
	if entity.ReleasedAt.Valid {
		t := entity.ReleasedAt.Time
		releasedAt = &t
	}

	return &Hold{
		ID:            entity.ID.String(),
		AccountID:     entity.AccountID.String(),
		TransactionID: entity.TransactionID.String(),
		Amount:        money.FromInt(entity.Amount),
		Description:   entity.Description.String,
		Status:        HoldStatus(entity.Status),
		CreatedAt:     entity.CreatedAt.Time,
		ReleasedAt:    releasedAt,
	}
}

func (h *Hold) ToSQLC() *sqlc.Hold {
	releasedAt := pgtype.Timestamptz{}
	if h.ReleasedAt != nil {
		releasedAt = pgtype.Timestamptz{Time: *h.ReleasedAt, Valid: true}
	}

	return &sqlc.Hold{
		ID:            utilid.FromString(h.ID).AsPgUUID(),
		AccountID:     utilid.FromString(h.AccountID).AsPgUUID(),
		TransactionID: utilid.FromString(h.TransactionID).AsPgUUID(),
		Amount:        h.Amount.ToInt(),
		Description:   pgtype.Text{String: h.Description, Valid: true},
		Status:        sqlc.HoldStatus(h.Status),
		CreatedAt:     pgtype.Timestamptz{Time: h.CreatedAt, Valid: true},
		ReleasedAt:    releasedAt,
	}
}
