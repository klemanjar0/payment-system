package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/pkg/money"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	"github.com/klemanjar0/payment-system/services/account/events"
)

type CreditUseCaseParams struct {
	AccountID, TransactionID string
	Amount                   money.Amount
}

type CreditUseCase struct {
	manager        domain.TransactionManager
	eventPublisher EventPublisher
}

func NewCreditUseCase(
	manager domain.TransactionManager,
	eventPublisher EventPublisher,
) *CreditUseCase {
	return &CreditUseCase{
		manager:        manager,
		eventPublisher: eventPublisher,
	}
}

func (uc *CreditUseCase) Execute(
	ctx context.Context,
	params *CreditUseCaseParams,
) error {
	err := uc.manager.ExecTx(ctx, func(tx domain.TxRepositories) error {
		// idempotency: skip if credit operation already recorded for this transaction
		_, err := tx.Operations.GetByTransactionIDAndType(ctx, params.AccountID, params.TransactionID, domain.OperationTypeCredit)
		if err == nil {
			logger.Info("Credit: already processed, skipping", "transaction_id", params.TransactionID)
			return nil
		}
		if !errors.Is(err, domain.ErrOperationNotFound) {
			return err
		}

		acc, err := tx.Accounts.GetByIDForUpdate(ctx, params.AccountID)
		if err != nil {
			return err
		}

		if err = acc.Credit(params.Amount); err != nil {
			return err
		}

		if err = tx.Accounts.Update(ctx, acc); err != nil {
			return err
		}

		return tx.Operations.Create(ctx, &domain.Operation{
			AccountID:     acc.ID,
			Type:          domain.OperationTypeCredit,
			Amount:        params.Amount.ToInt(),
			BalanceAfter:  acc.Balance.ToInt(),
			TransactionID: params.TransactionID,
			Description:   "system",
			CreatedAt:     time.Now(),
		})
	})

	if err != nil {
		logger.Error("Credit failed", "err", err, "account_id", params.AccountID, "transaction_id", params.TransactionID)
		return err
	}

	logger.Info("account credited", "account_id", params.AccountID, "transaction_id", params.TransactionID, "amount", params.Amount)

	if err = uc.eventPublisher.Publish(ctx, events.Credited, events.CreditedPayload{
		AccountID:     params.AccountID,
		TransactionID: params.TransactionID,
		Amount:        params.Amount.ToInt(),
	}); err != nil {
		logger.Error("Credit: failed to publish event", "err", err, "transaction_id", params.TransactionID)
		return err
	}

	return nil
}
