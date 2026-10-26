package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func StartOrderConsumer(ctx context.Context) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "order-events",
		GroupID:  "inventory-service-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	defer reader.Close()

	fmt.Println("Inventory Kafka consumer started")

	for {
		message, err := reader.ReadMessage(ctx)
		if err != nil {
			fmt.Println("Kafka read error:", err)
			return
		}

		fmt.Printf(
			"Received order event: %s\n",
			string(message.Value),
		)
	}
}
