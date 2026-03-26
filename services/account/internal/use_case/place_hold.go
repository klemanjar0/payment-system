package usecase

import (
	"context"
	"errors"

	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/pkg/money"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	"github.com/klemanjar0/payment-system/services/account/internal/events"
)

type PlaceHoldUseCaseParams struct {
	AccountID, TransactionID string
	Amount                   money.Amount
}

type PlaceHoldUseCase struct {
	manager        domain.TransactionManager
	eventPublisher EventPublisher
}

func NewPlaceHoldUseCase(
	manager domain.TransactionManager,
	eventPublisher EventPublisher,
) *PlaceHoldUseCase {
	return &PlaceHoldUseCase{
		manager:        manager,
		eventPublisher: eventPublisher,
	}
}

func (uc *PlaceHoldUseCase) Execute(
	ctx context.Context,
	params *PlaceHoldUseCaseParams,
) error {
	var amount money.Amount

	err := uc.manager.ExecTx(ctx, func(tx domain.TxRepositories) error {
		// idempotency: skip if hold already exists for this transaction
		existing, err := tx.Holds.GetByTransactionID(ctx, params.AccountID, params.TransactionID)
		if err == nil {
			logger.Info("PlaceHold: already processed, skipping", "transaction_id", params.TransactionID)
			amount = existing.Amount
			return nil
		}
		if !errors.Is(err, domain.ErrHoldNotFound) {
			return err
		}

		account, err := tx.Accounts.GetByIDForUpdate(ctx, params.AccountID)
		if err != nil {
			return err
		}

		if err = account.Hold(params.Amount); err != nil {
			return err
		}

		if err = tx.Accounts.Update(ctx, account); err != nil {
			return err
		}

		amount = params.Amount
		return tx.Holds.Create(ctx, domain.NewHold(
			params.AccountID,
			params.TransactionID,
			params.Amount,
			"system",
		))
	})

	if err != nil {
		logger.Error("PlaceHold failed", "err", err, "account_id", params.AccountID, "transaction_id", params.TransactionID)
		return err
	}

	logger.Info("hold placed", "account_id", params.AccountID, "transaction_id", params.TransactionID, "amount", amount)

	if err = uc.eventPublisher.Publish(ctx, events.HoldPlaced, events.HoldPlacedPayload{
		AccountID:     params.AccountID,
		TransactionID: params.TransactionID,
		Amount:        amount.ToInt(),
	}); err != nil {
		logger.Error("PlaceHold: failed to publish event", "err", err, "transaction_id", params.TransactionID)
		return err
	}

	return nil
}
