package domain

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/klemanjar0/payment-system/pkg/money"
	utilid "github.com/klemanjar0/payment-system/pkg/util_id"
	"github.com/klemanjar0/payment-system/services/account/internal/repository/postgres/sqlc"
)

type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "active"
	AccountStatusBlocked AccountStatus = "blocked"
	AccountStatusClosed  AccountStatus = "closed"
)

type Currency string

const (
	CurrencyUAH Currency = "UAH"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

func IsValidCurrency(c string) bool {
	switch Currency(c) {
	case CurrencyUAH, CurrencyUSD, CurrencyEUR:
		return true
	}
	return false
}

type Account struct {
	ID         string
	UserID     string
	Currency   Currency
	Balance    money.Amount
	HoldAmount money.Amount
	Status     AccountStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewAccount(userID string, currency Currency) *Account {
	now := time.Now()
	return &Account{
		UserID:     userID,
		Currency:   currency,
		Balance:    0,
		HoldAmount: 0,
		Status:     AccountStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func AccountFromSQLC(entity *sqlc.Account) (*Account, error) {
	if isValid := IsValidCurrency(entity.Currency); !isValid {
		return nil, ErrInvalidCurrency
	}

	return &Account{
		ID:         entity.ID.String(),
		UserID:     entity.UserID.String(),
		Currency:   Currency(entity.Currency),
		Balance:    money.FromInt(entity.Balance),
		HoldAmount: money.FromInt(entity.HoldAmount),
		Status:     AccountStatus(entity.Status),
		CreatedAt:  entity.CreatedAt.Time,
		UpdatedAt:  entity.UpdatedAt.Time,
	}, nil
}

// [requireId] = params;
func (a *Account) ToSQLC(params ...bool) (*sqlc.Account, error) {
	requireID := len(params) > 0 && params[0]
	if isValid := IsValidCurrency(string(a.Currency)); !isValid {
		return nil, ErrInvalidCurrency
	}

	id := utilid.FromString(a.ID).AsPgUUID()

	if requireID && !id.Valid {
		return nil, ErrAccountNullID
	}

	return &sqlc.Account{
		ID:         id,
		UserID:     utilid.FromString(a.UserID).AsPgUUID(),
		Currency:   string(a.Currency),
		Balance:    a.Balance.ToInt(),
		HoldAmount: a.HoldAmount.ToInt(),
		Status:     sqlc.AccountStatus(a.Status),
		CreatedAt:  pgtype.Timestamptz{Time: a.CreatedAt, Valid: true},
		UpdatedAt:  pgtype.Timestamptz{Time: a.UpdatedAt, Valid: true},
	}, nil
}

func (a *Account) Available() money.Amount {
	return a.Balance.Sub(a.HoldAmount)
}

func (a *Account) IsActive() bool {
	return a.Status == AccountStatusActive
}

func (a *Account) CanDebit(amount money.Amount) bool {
	return a.IsActive() && a.Available().GreaterThanOrEqual(amount)
}

func (a *Account) CanHold(amount money.Amount) bool {
	return a.IsActive() && a.Available().GreaterThanOrEqual(amount)
}

func (a *Account) Hold(amount money.Amount) error {
	if !a.IsActive() {
		return ErrAccountNotActive
	}
	if !a.Available().GreaterThanOrEqual(amount) {
		return ErrInsufficientFunds
	}

	a.HoldAmount = a.HoldAmount.Add(amount)
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) ReleaseHold(amount money.Amount) error {
	if !a.HoldAmount.GreaterThanOrEqual(amount) {
		return ErrInvalidHoldAmount
	}

	a.HoldAmount = a.HoldAmount.Sub(amount)
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Credit(amount money.Amount) error {
	if !a.IsActive() {
		return ErrAccountNotActive
	}
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}

	a.Balance = a.Balance.Add(amount)
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Debit(amount money.Amount) error {
	if !a.IsActive() {
		return ErrAccountNotActive
	}
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	if !a.Available().GreaterThanOrEqual(amount) {
		return ErrInsufficientFunds
	}

	a.Balance = a.Balance.Sub(amount)
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) DebitFromHold(amount money.Amount) error {
	if !a.IsActive() {
		return ErrAccountNotActive
	}
	if !a.HoldAmount.GreaterThanOrEqual(amount) {
		return ErrInvalidHoldAmount
	}

	a.HoldAmount = a.HoldAmount.Sub(amount)
	a.Balance = a.Balance.Sub(amount)
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Block() {
	a.Status = AccountStatusBlocked
	a.UpdatedAt = time.Now()
}

func (a *Account) Close() error {
	if a.Balance != 0 {
		return ErrAccountHasBalance
	}
	if a.HoldAmount != 0 {
		return ErrAccountHasHolds
	}

	a.Status = AccountStatusClosed
	a.UpdatedAt = time.Now()
	return nil
}
