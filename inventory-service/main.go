package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/go-commerce-platform/inventory-service/config"
	"example.com/go-commerce-platform/inventory-service/handler"
	inventoryKafka "example.com/go-commerce-platform/inventory-service/kafka"
	"example.com/go-commerce-platform/inventory-service/repository"
	"example.com/go-commerce-platform/inventory-service/service"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("failed to load inventory-service .env")
	}

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := repository.NewInventoryRepository(db)
	inventoryService := service.NewInventoryService(repo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	mux := http.NewServeMux()
	mux.HandleFunc("/products", inventoryHandler.Products)
	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= len("/products/stock") &&
			r.URL.Path[len(r.URL.Path)-6:] == "/stock" {
			inventoryHandler.Stock(w, r)
			return
		}

		inventoryHandler.ProductByID(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Start Kafka consumer in a separate goroutine.
	go inventoryKafka.StartOrderConsumer(ctx)

	// Gracefully shut down HTTP server when context is cancelled.
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("Inventory Service starting on port %s", port)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
