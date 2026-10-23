package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafkago.Reader
}

func NewConsumer() *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "order-events",
		GroupID: "order-events-consumer-group",
	})

	return &Consumer{
		reader: reader,
	}
}

func (c *Consumer) ConsumeOrderEvents(ctx context.Context) error {
	for {
		message, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("failed to read Kafka message: %w", err)
		}

		var event OrderCreatedEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			fmt.Printf("Invalid order event: %v\n", err)
			continue
		}

		fmt.Printf(
			"Order event received: event=%s order_id=%d user_id=%d amount=%.2f\n",
			event.Event,
			event.OrderID,
			event.UserID,
			event.Amount,
		)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
