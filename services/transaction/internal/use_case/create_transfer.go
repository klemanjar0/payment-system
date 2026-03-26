package usecase

import (
	"context"
	"errors"

	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/pkg/money"
	"github.com/klemanjar0/payment-system/services/transaction/events"
	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
)

type CreateTransferUseCase struct {
	repo          domain.TransactionRepository
	accountClient AccountServiceClient
	userClient    UserServiceClient
	cmdPublisher  CommandPublisher
}

func NewCreateTransferUseCase(
	repo domain.TransactionRepository,
	accountClient AccountServiceClient,
	userClient UserServiceClient,
	cmdPublisher CommandPublisher,
) *CreateTransferUseCase {
	return &CreateTransferUseCase{
		repo:          repo,
		accountClient: accountClient,
		userClient:    userClient,
		cmdPublisher:  cmdPublisher,
	}
}

func (uc *CreateTransferUseCase) Execute(ctx context.Context, params *CreateTransferParams) (*domain.Transaction, error) {
	// Idempotency: return existing transaction if already created.
	if params.IdempotencyKey != "" {
		existing, err := uc.repo.GetByIdempotencyKey(ctx, params.IdempotencyKey)
		if err == nil {
			logger.Info("CreateTransfer: idempotent, returning existing", "idempotency_key", params.IdempotencyKey)
			return existing, nil
		}
		if !errors.Is(err, domain.ErrTransactionNotFound) {
			return nil, err
		}
	}

	if params.FromAccountID == params.ToAccountID {
		return nil, domain.ErrSameAccount
	}
	if params.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	fromAccount, err := uc.accountClient.GetAccount(ctx, params.FromAccountID)
	if err != nil {
		return nil, err
	}
	if fromAccount.Status != "active" {
		return nil, domain.ErrAccountNotActive
	}

	toAccount, err := uc.accountClient.GetAccount(ctx, params.ToAccountID)
	if err != nil {
		return nil, err
	}
	if toAccount.Status != "active" {
		return nil, domain.ErrAccountNotActive
	}

	if fromAccount.Currency != toAccount.Currency {
		return nil, domain.ErrCurrencyMismatch
	}

	user, err := uc.userClient.GetUser(ctx, fromAccount.UserID)
	if err != nil {
		return nil, err
	}
	if user.Status != "active" {
		return nil, domain.ErrUserNotActive
	}

	tx := domain.NewTransaction(
		params.IdempotencyKey,
		params.FromAccountID,
		params.ToAccountID,
		money.FromInt(params.Amount),
		fromAccount.Currency,
		params.Description,
	)

	if err = uc.repo.Create(ctx, tx); err != nil {
		logger.Error("CreateTransfer: failed to persist transaction", "err", err)
		return nil, err
	}

	logger.Info("transaction created", "id", tx.ID, "from", params.FromAccountID, "to", params.ToAccountID, "amount", params.Amount)

	// Kick off the saga: request a hold on sender account.
	if err = uc.cmdPublisher.Publish(ctx, events.HoldRequested, events.HoldRequestedPayload{
		AccountID:     params.FromAccountID,
		TransactionID: tx.ID,
		Amount:        params.Amount,
	}); err != nil {
		logger.Error("CreateTransfer: failed to publish hold_requested", "err", err, "transaction_id", tx.ID)
		return nil, err
	}

	return tx, nil
}
