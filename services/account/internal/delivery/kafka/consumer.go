package kafka

import (
	"context"
	"encoding/json"

	pkgkafka "github.com/klemanjar0/payment-system/pkg/kafka"
	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/pkg/money"
	usecase "github.com/klemanjar0/payment-system/services/account/internal/use_case"
	txevents "github.com/klemanjar0/payment-system/services/transaction/events"
)

type SagaConsumer struct {
	consumer     *pkgkafka.Consumer
	holdPlacer   usecase.HoldPlacer
	holdExecutor usecase.HoldExecutor
	holdReleaser usecase.HoldReleaser
	creditor     usecase.AccountCreditor
}

func NewSagaConsumer(
	consumer *pkgkafka.Consumer,
	holdPlacer usecase.HoldPlacer,
	holdExecutor usecase.HoldExecutor,
	holdReleaser usecase.HoldReleaser,
	creditor usecase.AccountCreditor,
) *SagaConsumer {
	return &SagaConsumer{
		consumer:     consumer,
		holdPlacer:   holdPlacer,
		holdExecutor: holdExecutor,
		holdReleaser: holdReleaser,
		creditor:     creditor,
	}
}

func (c *SagaConsumer) Run(ctx context.Context) {
	logger.Info("SagaConsumer started")
	c.consumer.Consume(ctx, func(ctx context.Context, event pkgkafka.Event) error {
		logger.Info("saga event received", "type", event.Type)
		if err := c.dispatch(ctx, event); err != nil {
			logger.Error("saga event failed", "type", event.Type, "err", err)
			return err
		}
		return nil
	})
}

func (c *SagaConsumer) dispatch(ctx context.Context, event pkgkafka.Event) error {
	raw, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	switch event.Type {
	case txevents.HoldRequested:
		var p txevents.HoldRequestedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		logger.Info("placing hold", "account_id", p.AccountID, "transaction_id", p.TransactionID, "amount", p.Amount)
		return c.holdPlacer.Execute(ctx, &usecase.PlaceHoldUseCaseParams{
			AccountID:     p.AccountID,
			TransactionID: p.TransactionID,
			Amount:        money.FromInt(p.Amount),
		})

	case txevents.ExecuteHoldRequested:
		var p txevents.ExecuteHoldRequestedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		logger.Info("executing hold", "account_id", p.AccountID, "transaction_id", p.TransactionID)
		return c.holdExecutor.Execute(ctx, &usecase.ExecuteHoldUseCaseParams{
			AccountID:     p.AccountID,
			TransactionID: p.TransactionID,
		})

	case txevents.ReleaseHoldRequested:
		var p txevents.ReleaseHoldRequestedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		logger.Info("releasing hold", "account_id", p.AccountID, "transaction_id", p.TransactionID)
		return c.holdReleaser.Execute(ctx, &usecase.ReleaseHoldUseCaseParams{
			AccountID:     p.AccountID,
			TransactionID: p.TransactionID,
		})

	case txevents.CreditRequested:
		var p txevents.CreditRequestedPayload
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		logger.Info("crediting account", "account_id", p.AccountID, "transaction_id", p.TransactionID, "amount", p.Amount)
		return c.creditor.Execute(ctx, &usecase.CreditUseCaseParams{
			AccountID:     p.AccountID,
			TransactionID: p.TransactionID,
			Amount:        money.FromInt(p.Amount),
		})

	default:
		logger.Info("unknown event type, skipping", "type", event.Type)
		return nil
	}
}
