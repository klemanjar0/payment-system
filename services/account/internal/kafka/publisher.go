package kafka

import (
	"context"
	"time"

	pkgkafka "github.com/klemanjar0/payment-system/pkg/kafka"
)

type Publisher struct {
	producer *pkgkafka.Producer
}

func NewPublisher(producer *pkgkafka.Producer) *Publisher {
	return &Publisher{producer: producer}
}

func (p *Publisher) Publish(ctx context.Context, eventType string, payload interface{}) error {
	return p.producer.Publish(ctx, eventType, pkgkafka.Event{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	})
}
