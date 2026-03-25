package domain

import "errors"

var (
	ErrAccountNullID     = errors.New("account id miss")
	ErrAccountNotFound   = errors.New("account not found")
	ErrAccountNotActive  = errors.New("account is not active")
	ErrAccountExists     = errors.New("account already exists for this currency")
	ErrAccountHasBalance = errors.New("account has non-zero balance")
	ErrAccountHasHolds   = errors.New("account has active holds")

	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrInvalidCurrency   = errors.New("invalid currency")
	ErrInvalidHoldAmount = errors.New("invalid hold amount")

	ErrHoldNotFound      = errors.New("hold not found")
	ErrHoldNotActive     = errors.New("hold is not active")
	ErrHoldAlreadyExists = errors.New("hold already exists for this transaction")

	ErrInternal = errors.New("internal error")
)
