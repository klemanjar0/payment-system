package usecase

import (
	"context"
	"errors"

	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
)

type CreateAccountUseCaseParams struct {
	UserID   string
	Currency domain.Currency
}

type CreateAccountUseCase struct {
	accounts domain.AccountRepository
}

func NewCreateAccountUseCase(accounts domain.AccountRepository) *CreateAccountUseCase {
	return &CreateAccountUseCase{accounts: accounts}
}

func (uc *CreateAccountUseCase) Execute(
	ctx context.Context,
	params *CreateAccountUseCaseParams,
) (*domain.Account, error) {
	if !domain.IsValidCurrency(string(params.Currency)) {
		return nil, domain.ErrInvalidCurrency
	}

	_, err := uc.accounts.GetByUserAndCurrency(ctx, params.UserID, params.Currency)
	if err == nil {
		return nil, domain.ErrAccountExists
	}
	if !errors.Is(err, domain.ErrAccountNotFound) {
		logger.Error("CreateAccount: failed to check existing account", "err", err, "user_id", params.UserID)
		return nil, err
	}

	account := domain.NewAccount(params.UserID, params.Currency)
	if err = uc.accounts.Create(ctx, account); err != nil {
		logger.Error("CreateAccount: failed to create account", "err", err, "user_id", params.UserID)
		return nil, err
	}

	logger.Info("account created", "account_id", account.ID, "user_id", params.UserID, "currency", params.Currency)
	return account, nil
}
