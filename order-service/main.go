package main

import (
	"fmt"
	"log"
	"net/http"

	"example.com/go-commerce-platform/order-service/config"
	"example.com/go-commerce-platform/order-service/handler"
	"example.com/go-commerce-platform/order-service/repository"
	"example.com/go-commerce-platform/order-service/route"
	"example.com/go-commerce-platform/order-service/service"
)

func main() {
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo)
	orderHandler := handler.NewOrderHandler(orderService)
	route.RegisterOrderRoutes(orderHandler)

	fmt.Println("Server started and running on port 8081")
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}

}
