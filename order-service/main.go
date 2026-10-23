package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/go-commerce-platform/order-service/config"
	"example.com/go-commerce-platform/order-service/handler"
	"example.com/go-commerce-platform/order-service/kafka"
	"example.com/go-commerce-platform/order-service/repository"
	"example.com/go-commerce-platform/order-service/route"
	"example.com/go-commerce-platform/order-service/service"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables.
	err := godotenv.Load("../user-service/.env")
	if err != nil {
		log.Fatal("Error loading user-service .env file")
	}

	// Connect to the database.
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create Kafka producer.
	producer := kafka.NewProducer()
	defer producer.Close()

	// Create repository, service, and handler.
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, producer)
	orderHandler := handler.NewOrderHandler(orderService)

	// Register existing order routes.
	route.RegisterOrderRoutes(orderHandler)

	// Handle application shutdown.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Start Kafka consumer.
	consumer := kafka.NewConsumer()
	defer consumer.Close()

	go func() {
		if err := consumer.ConsumeOrderEvents(ctx); err != nil {
			log.Printf("Kafka consumer stopped: %v", err)
			stop()
		}
	}()

	// Create HTTP server.
	server := &http.Server{
		Addr:              ":8081",
		Handler:           http.DefaultServeMux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Shut down HTTP server when application stops.
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	fmt.Println("Order service started on port 8081")

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
