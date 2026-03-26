package usecase

import (
	"context"
	"time"

	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/pkg/money"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	"github.com/klemanjar0/payment-system/services/account/internal/events"
)

type ExecuteHoldUseCaseParams struct {
	AccountID, TransactionID, HoldID string
}

type ExecuteHoldUseCase struct {
	manager        domain.TransactionManager
	eventPublisher EventPublisher
}

func NewExecuteHoldUseCase(
	manager domain.TransactionManager,
	eventPublisher EventPublisher,
) *ExecuteHoldUseCase {
	return &ExecuteHoldUseCase{
		manager:        manager,
		eventPublisher: eventPublisher,
	}
}

func (uc *ExecuteHoldUseCase) Execute(
	ctx context.Context,
	params *ExecuteHoldUseCaseParams,
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

		// idempotency: already executed
		if hold.Status == domain.HoldStatusExecuted {
			logger.Info("ExecuteHold: already processed, skipping", "transaction_id", params.TransactionID)
			amount = hold.Amount
			return nil
		}

		amount = hold.Amount

		if err = acc.DebitFromHold(hold.Amount); err != nil {
			return err
		}
		hold.Execute()

		if err = tx.Holds.Update(ctx, hold); err != nil {
			return err
		}

		if err = tx.Accounts.Update(ctx, acc); err != nil {
			return err
		}

		return tx.Operations.Create(ctx, &domain.Operation{
			AccountID:     acc.ID,
			Type:          domain.OperationTypeDebit,
			Amount:        hold.Amount.ToInt(),
			BalanceAfter:  acc.Balance.ToInt(),
			TransactionID: params.TransactionID,
			Description:   "system",
			CreatedAt:     time.Now(),
		})
	})

	if err != nil {
		logger.Error("ExecuteHold failed", "err", err, "account_id", params.AccountID, "transaction_id", params.TransactionID)
		return err
	}

	logger.Info("hold executed", "account_id", params.AccountID, "transaction_id", params.TransactionID, "amount", amount)

	if err = uc.eventPublisher.Publish(ctx, events.HoldExecuted, events.HoldExecutedPayload{
		AccountID:     params.AccountID,
		TransactionID: params.TransactionID,
		Amount:        amount.ToInt(),
	}); err != nil {
		logger.Error("ExecuteHold: failed to publish event", "err", err, "transaction_id", params.TransactionID)
		return err
	}

	return nil
}
