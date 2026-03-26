package kafka

import (
	"context"
	"encoding/json"

	pkgkafka "github.com/klemanjar0/payment-system/pkg/kafka"
	"github.com/klemanjar0/payment-system/pkg/logger"
	accountevents "github.com/klemanjar0/payment-system/services/account/events"
	txevents "github.com/klemanjar0/payment-system/services/transaction/events"
	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
)

// SagaOrchestrator listens to account-events and drives the saga forward.
type SagaOrchestrator struct {
	consumer     *pkgkafka.Consumer
	txRepo       domain.TransactionRepository
	cmdPublisher *Publisher
}

func NewSagaOrchestrator(
	consumer *pkgkafka.Consumer,
	txRepo domain.TransactionRepository,
	cmdPublisher *Publisher,
) *SagaOrchestrator {
	return &SagaOrchestrator{
		consumer:     consumer,
		txRepo:       txRepo,
		cmdPublisher: cmdPublisher,
	}
}

func (o *SagaOrchestrator) Run(ctx context.Context) {
	logger.Info("SagaOrchestrator started")
	o.consumer.Consume(ctx, func(ctx context.Context, event pkgkafka.Event) error {
		logger.Info("saga orchestrator received event", "type", event.Type)
		if err := o.dispatch(ctx, event); err != nil {
			logger.Error("saga orchestrator failed to process event", "type", event.Type, "err", err)
			return err
		}
		return nil
	})
}

func (o *SagaOrchestrator) dispatch(ctx context.Context, event pkgkafka.Event) error {
	raw, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	switch event.Type {
	case accountevents.HoldPlaced:
		var p accountevents.HoldPlacedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		return o.onHoldPlaced(ctx, p)

	case accountevents.HoldExecuted:
		var p accountevents.HoldExecutedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		return o.onHoldExecuted(ctx, p)

	case accountevents.Credited:
		var p accountevents.CreditedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		return o.onCredited(ctx, p)

	case accountevents.HoldReleased:
		var p accountevents.HoldReleasedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		return o.onHoldReleased(ctx, p)

	case accountevents.HoldFailed:
		var p accountevents.HoldFailedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		return o.onHoldFailed(ctx, p)

	default:
		logger.Info("saga orchestrator: unknown event type, skipping", "type", event.Type)
		return nil
	}
}

func (o *SagaOrchestrator) onHoldPlaced(ctx context.Context, p accountevents.HoldPlacedPayload) error {
	tx, err := o.txRepo.GetByID(ctx, p.TransactionID)
	if err != nil {
		logger.Error("onHoldPlaced: transaction not found", "transaction_id", p.TransactionID, "err", err)
		return err
	}
	if tx.IsTerminal() {
		logger.Info("onHoldPlaced: transaction already terminal, skipping", "transaction_id", p.TransactionID, "status", tx.Status)
		return nil
	}

	logger.Info("hold placed, requesting execute_hold", "transaction_id", p.TransactionID)
	return o.cmdPublisher.Publish(ctx, txevents.ExecuteHoldRequested, txevents.ExecuteHoldRequestedPayload{
		AccountID:     p.AccountID,
		TransactionID: p.TransactionID,
	})
}

func (o *SagaOrchestrator) onHoldExecuted(ctx context.Context, p accountevents.HoldExecutedPayload) error {
	tx, err := o.txRepo.GetByID(ctx, p.TransactionID)
	if err != nil {
		logger.Error("onHoldExecuted: transaction not found", "transaction_id", p.TransactionID, "err", err)
		return err
	}
	if tx.IsTerminal() {
		logger.Info("onHoldExecuted: transaction already terminal, skipping", "transaction_id", p.TransactionID)
		return nil
	}

	if err = o.txRepo.UpdateStatus(ctx, tx.ID, domain.StatusProcessing); err != nil {
		logger.Error("onHoldExecuted: failed to update status", "err", err, "transaction_id", tx.ID)
		return err
	}

	logger.Info("hold executed, requesting credit", "transaction_id", p.TransactionID, "to_account", tx.ToAccountID)
	return o.cmdPublisher.Publish(ctx, txevents.CreditRequested, txevents.CreditRequestedPayload{
		AccountID:     tx.ToAccountID,
		TransactionID: tx.ID,
		Amount:        tx.Amount.ToInt(),
	})
}

func (o *SagaOrchestrator) onCredited(ctx context.Context, p accountevents.CreditedPayload) error {
	tx, err := o.txRepo.GetByID(ctx, p.TransactionID)
	if err != nil {
		logger.Error("onCredited: transaction not found", "transaction_id", p.TransactionID, "err", err)
		return err
	}
	if tx.IsTerminal() {
		logger.Info("onCredited: transaction already terminal, skipping", "transaction_id", p.TransactionID)
		return nil
	}

	if err = o.txRepo.UpdateStatus(ctx, tx.ID, domain.StatusCompleted); err != nil {
		logger.Error("onCredited: failed to update status", "err", err, "transaction_id", tx.ID)
		return err
	}

	logger.Info("transaction completed", "transaction_id", tx.ID, "from", tx.FromAccountID, "to", tx.ToAccountID)
	return nil
}

func (o *SagaOrchestrator) onHoldReleased(ctx context.Context, p accountevents.HoldReleasedPayload) error {
	tx, err := o.txRepo.GetByID(ctx, p.TransactionID)
	if err != nil {
		logger.Error("onHoldReleased: transaction not found", "transaction_id", p.TransactionID, "err", err)
		return err
	}
	if tx.IsTerminal() {
		logger.Info("onHoldReleased: transaction already terminal, skipping", "transaction_id", p.TransactionID)
		return nil
	}

	if err = o.txRepo.UpdateStatusWithReason(ctx, tx.ID, domain.StatusFailed, "hold released (compensated)"); err != nil {
		logger.Error("onHoldReleased: failed to update status", "err", err, "transaction_id", tx.ID)
		return err
	}

	logger.Info("transaction failed (compensated)", "transaction_id", tx.ID)
	return nil
}

func (o *SagaOrchestrator) onHoldFailed(ctx context.Context, p accountevents.HoldFailedPayload) error {
	tx, err := o.txRepo.GetByID(ctx, p.TransactionID)
	if err != nil {
		logger.Error("onHoldFailed: transaction not found", "transaction_id", p.TransactionID, "err", err)
		return err
	}
	if tx.IsTerminal() {
		logger.Info("onHoldFailed: transaction already terminal, skipping", "transaction_id", p.TransactionID)
		return nil
	}

	if err = o.txRepo.UpdateStatusWithReason(ctx, tx.ID, domain.StatusFailed, p.Reason); err != nil {
		logger.Error("onHoldFailed: failed to update status", "err", err, "transaction_id", tx.ID)
		return err
	}

	logger.Info("transaction failed (hold rejected)", "transaction_id", tx.ID, "reason", p.Reason)
	return nil
}
