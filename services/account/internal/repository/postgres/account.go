package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/klemanjar0/payment-system/pkg/logger"
	utilid "github.com/klemanjar0/payment-system/pkg/util_id"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	"github.com/klemanjar0/payment-system/services/account/internal/repository/postgres/sqlc"
)

type AccountRepository struct {
	queries *sqlc.Queries
}

func NewAccountRepository(queries *sqlc.Queries) *AccountRepository {
	return &AccountRepository{
		queries: queries,
	}
}

func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	pgTypedAccount, mapError := account.ToSQLC(false)

	if mapError != nil {
		return mapError
	}

	_, err := r.queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		UserID:     pgTypedAccount.UserID,
		Currency:   pgTypedAccount.Currency,
		Balance:    pgTypedAccount.Balance,
		HoldAmount: pgTypedAccount.HoldAmount,
		Status:     pgTypedAccount.Status,
		CreatedAt:  pgTypedAccount.CreatedAt,
		UpdatedAt:  pgTypedAccount.UpdatedAt,
	})
	return err
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	row, err := r.queries.GetAccountByID(ctx, utilid.FromString(id).AsPgUUID())

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}

	if err != nil {
		return nil, err
	}

	account, mapError := domain.AccountFromSQLC(&row)

	if mapError != nil {
		logger.Error("failed to map account", "err", mapError, "module", "AccountRepository:GetByID")
		return nil, mapError
	}

	return account, nil
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Account, error) {
	rows, err := r.queries.GetAccountsByUserID(ctx, utilid.FromString(userID).AsPgUUID())
	if err != nil {
		return nil, err
	}

	accounts := make([]*domain.Account, len(rows))
	for i, row := range rows {
		accounts[i], err = domain.AccountFromSQLC(&row)
		if err != nil {
			logger.Error("failed to map account", "err", err, "module", "AccountRepository:GetByUserID")
			return nil, err
		}
	}
	return accounts, nil
}

func (r *AccountRepository) GetByUserAndCurrency(ctx context.Context, userID string, currency domain.Currency) (*domain.Account, error) {
	row, err := r.queries.GetAccountByUserAndCurrency(ctx, sqlc.GetAccountByUserAndCurrencyParams{
		UserID:   utilid.FromString(userID).AsPgUUID(),
		Currency: string(currency),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}

	account, mapError := domain.AccountFromSQLC(&row)

	if mapError != nil {
		logger.Error("failed to map account", "err", mapError, "module", "AccountRepository:GetByUserAndCurrency")
		return nil, mapError
	}

	return account, nil
}

func (r *AccountRepository) GetByIDForUpdate(ctx context.Context, id string) (*domain.Account, error) {
	row, err := r.queries.GetAccountByIDForUpdate(ctx, utilid.FromString(id).AsPgUUID())

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}

	if err != nil {
		return nil, err
	}

	account, mapError := domain.AccountFromSQLC(&row)

	if mapError != nil {
		logger.Error("failed to map account", "err", mapError, "module", "AccountRepository:GetByID")
		return nil, mapError
	}

	return account, nil
}

func (r *AccountRepository) Update(ctx context.Context, account *domain.Account) error {
	pgTypedAccount, mapError := account.ToSQLC(false)

	if mapError != nil {
		logger.Error("failed to map account", "err", mapError, "module", "AccountRepository:Update", "reason", "empty account id found")
		return mapError
	}

	_, err := r.queries.UpdateAccount(ctx, sqlc.UpdateAccountParams{
		ID:         pgTypedAccount.ID,
		Balance:    pgTypedAccount.Balance,
		HoldAmount: pgTypedAccount.HoldAmount,
		Status:     pgTypedAccount.Status,
		UpdatedAt:  pgTypedAccount.UpdatedAt,
	})

	return err
}
