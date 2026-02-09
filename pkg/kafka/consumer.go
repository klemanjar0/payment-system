package kafka

import (
	"context"
	"encoding/json"

	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewConsumer(cfg ConsumerConfig) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{reader: reader}
}

type MessageHandler func(ctx context.Context, event Event) error

func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				logger.Error("failed to read message", "error", err)
				continue
			}

			var event Event
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				logger.Error("failed to unmarshal event", "error", err)
				continue
			}

			if err := handler(ctx, event); err != nil {
				logger.Error("failed to handle event", "error", err, "type", event.Type)
				// add retry logic or dead letter queue for future
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
