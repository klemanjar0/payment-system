package usecase

import (
	"context"
	"time"

	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/pkg/money"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	"github.com/klemanjar0/payment-system/services/account/internal/events"
)

type ReleaseHoldUseCaseParams struct {
	AccountID, TransactionID string
}

type ReleaseHoldUseCase struct {
	manager        domain.TransactionManager
	eventPublisher EventPublisher
}

func NewReleaseHoldUseCase(
	manager domain.TransactionManager,
	eventPublisher EventPublisher,
) *ReleaseHoldUseCase {
	return &ReleaseHoldUseCase{
		manager:        manager,
		eventPublisher: eventPublisher,
	}
}

func (uc *ReleaseHoldUseCase) Execute(
	ctx context.Context,
	params *ReleaseHoldUseCaseParams,
) error {
	var amount money.Amount

	err := uc.manager.ExecTx(ctx, func(tx domain.TxRepositories) error {
		acc, err := tx.Accounts.GetByIDForUpdate(ctx, params.AccountID)
		if err != nil {
			return err
		}

		hold, err := tx.Holds.GetByTransactionIDForUpdate(ctx, params.AccountID, params.TransactionID)
		if err != nil {
			return err
		}

		// idempotency: already released
		if hold.Status == domain.HoldStatusReleased {
			logger.Info("ReleaseHold: already processed, skipping", "transaction_id", params.TransactionID)
			amount = hold.Amount
			return nil
		}

		amount = hold.Amount

		if err = acc.ReleaseHold(hold.Amount); err != nil {
			return err
		}
		hold.Release()

		if err = tx.Holds.Update(ctx, hold); err != nil {
			return err
		}

		if err = tx.Accounts.Update(ctx, acc); err != nil {
			return err
		}

		return tx.Operations.Create(ctx, &domain.Operation{
			AccountID:     acc.ID,
			Type:          domain.OperationTypeHoldRelease,
			Amount:        hold.Amount.ToInt(),
			BalanceAfter:  acc.Balance.ToInt(),
			TransactionID: params.TransactionID,
			Description:   "system",
			CreatedAt:     time.Now(),
		})
	})

	if err != nil {
		logger.Error("ReleaseHold failed", "err", err, "account_id", params.AccountID, "transaction_id", params.TransactionID)
		return err
	}

	logger.Info("hold released", "account_id", params.AccountID, "transaction_id", params.TransactionID, "amount", amount)

	if err = uc.eventPublisher.Publish(ctx, events.HoldReleased, events.HoldReleasedPayload{
		AccountID:     params.AccountID,
		TransactionID: params.TransactionID,
		Amount:        amount.ToInt(),
	}); err != nil {
		logger.Error("ReleaseHold: failed to publish event", "err", err, "transaction_id", params.TransactionID)
		return err
	}

	return nil
}
