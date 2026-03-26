package domain

import "errors"

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrTransactionExists   = errors.New("transaction already exists")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInvalidCurrency     = errors.New("invalid currency")
	ErrAccountNotFound     = errors.New("account not found")
	ErrAccountNotActive    = errors.New("account not active")
	ErrCurrencyMismatch    = errors.New("currency mismatch between accounts")
	ErrSameAccount         = errors.New("source and destination accounts are the same")
	ErrUserNotActive       = errors.New("user is not active")
)
