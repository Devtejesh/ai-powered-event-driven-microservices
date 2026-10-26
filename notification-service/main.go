package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/segmentio/kafka-go"
)

type OrderCreatedEvent struct {
	Event   string  `json:"event"`
	OrderID int     `json:"order_id"`
	UserID  int     `json:"user_id"`
	Amount  float64 `json:"amount"`
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(), os.Interrupt,
	)
	defer stop()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "order-events",
		GroupID: "notification-service-group",
	})
	defer reader.Close()

	fmt.Println("Notification Service started...")

	for {
		message, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("Kafka read error: %v", err)
			continue
		}

		var event OrderCreatedEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("Invalid event: %v", err)
			continue
		}

		if event.Event == "OrderCreated" {
			fmt.Printf(
				"NOTIFICATION: Order %d created for User %d. Amount: %.2f\n",
				event.OrderID,
				event.UserID,
				event.Amount,
			)
		}
	}
}
