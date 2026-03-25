package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	utilid "github.com/klemanjar0/payment-system/pkg/util_id"
	"github.com/klemanjar0/payment-system/services/account/internal/repository/postgres/sqlc"
)

type OperationType string

const (
	OperationTypeCredit      OperationType = "credit"
	OperationTypeDebit       OperationType = "debit"
	OperationTypeHold        OperationType = "hold"
	OperationTypeHoldRelease OperationType = "hold_release"
)

type Operation struct {
	ID            string
	AccountID     string
	Type          OperationType
	Amount        int64
	BalanceAfter  int64
	TransactionID string
	Description   string
	CreatedAt     time.Time
}

func NewOperation(
	accountID string,
	opType OperationType,
	amount int64,
	balanceAfter int64,
	transactionID string,
	description string,
) *Operation {
	return &Operation{
		AccountID:     accountID,
		Type:          opType,
		Amount:        amount,
		BalanceAfter:  balanceAfter,
		TransactionID: transactionID,
		Description:   description,
		CreatedAt:     time.Now(),
	}
}

func OperationFromSQLC(entity *sqlc.Operation) *Operation {
	return &Operation{
		ID:            entity.ID.String(),
		AccountID:     entity.AccountID.String(),
		Type:          OperationType(entity.Type),
		Amount:        entity.Amount,
		BalanceAfter:  entity.BalanceAfter,
		TransactionID: entity.TransactionID.String(),
		Description:   entity.Description.String,
		CreatedAt:     entity.CreatedAt.Time,
	}
}

func (o *Operation) ToSQLC() *sqlc.Operation {
	return &sqlc.Operation{
		ID:            utilid.FromString(o.ID).AsPgUUID(),
		AccountID:     utilid.FromString(o.AccountID).AsPgUUID(),
		Type:          sqlc.OperationType(o.Type),
		Amount:        o.Amount,
		BalanceAfter:  o.BalanceAfter,
		TransactionID: utilid.FromString(o.TransactionID).AsPgUUID(),
		Description:   pgtype.Text{String: o.Description, Valid: true},
		CreatedAt:     pgtype.Timestamptz{Time: o.CreatedAt, Valid: true},
	}
}
