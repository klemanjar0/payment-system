package domain

import (
	"time"

	"github.com/klemanjar0/payment-system/pkg/money"
	utilid "github.com/klemanjar0/payment-system/pkg/util_id"
	"github.com/klemanjar0/payment-system/services/transaction/internal/repository/postgres/sqlc"
)

type TransactionStatus string

const (
	StatusPending    TransactionStatus = "pending"
	StatusProcessing TransactionStatus = "processing"
	StatusCompleted  TransactionStatus = "completed"
	StatusFailed     TransactionStatus = "failed"
	StatusReversed   TransactionStatus = "reversed"
)

type Transaction struct {
	ID             string
	IdempotencyKey string
	FromAccountID  string
	ToAccountID    string
	Amount         money.Amount
	Currency       string
	Description    string
	Status         TransactionStatus
	FailureReason  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewTransaction(idempotencyKey, fromAccountID, toAccountID string, amount money.Amount, currency, description string) *Transaction {
	return &Transaction{
		IdempotencyKey: idempotencyKey,
		FromAccountID:  fromAccountID,
		ToAccountID:    toAccountID,
		Amount:         amount,
		Currency:       currency,
		Description:    description,
		Status:         StatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (t *Transaction) SetProcessing() {
	t.Status = StatusProcessing
	t.UpdatedAt = time.Now()
}

func (t *Transaction) Complete() {
	t.Status = StatusCompleted
	t.UpdatedAt = time.Now()
}

func (t *Transaction) Fail(reason string) {
	t.Status = StatusFailed
	t.FailureReason = reason
	t.UpdatedAt = time.Now()
}

func (t *Transaction) Reverse() {
	t.Status = StatusReversed
	t.UpdatedAt = time.Now()
}

func (t *Transaction) IsTerminal() bool {
	return t.Status == StatusCompleted || t.Status == StatusFailed || t.Status == StatusReversed
}

func (t *Transaction) ToSQLC() sqlc.CreateTransactionParams {
	return sqlc.CreateTransactionParams{
		IdempotencyKey: t.IdempotencyKey,
		FromAccountID:  utilid.FromString(t.FromAccountID).AsPgUUID(),
		ToAccountID:    utilid.FromString(t.ToAccountID).AsPgUUID(),
		Amount:         t.Amount.ToInt(),
		Currency:       t.Currency,
		Description:    t.Description,
		Status:         sqlc.TransactionStatus(t.Status),
	}
}

func TransactionFromSQLC(row sqlc.Transaction) *Transaction {
	return &Transaction{
		ID:             utilid.FromPg(row.ID).AsString(),
		IdempotencyKey: row.IdempotencyKey,
		FromAccountID:  utilid.FromPg(row.FromAccountID).AsString(),
		ToAccountID:    utilid.FromPg(row.ToAccountID).AsString(),
		Amount:         money.FromInt(row.Amount),
		Currency:       row.Currency,
		Description:    row.Description,
		Status:         TransactionStatus(row.Status),
		FailureReason:  row.FailureReason,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
}
