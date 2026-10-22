package route

import (
	"net/http"

	"example.com/go-commerce-platform/order-service/handler"
)

func RegisterOrderRoutes(orderHandler *handler.OrderHandler) {

	http.HandleFunc("POST /orders", orderHandler.CreateOrder)
	http.HandleFunc("GET /orders", orderHandler.ListOrders)
	http.HandleFunc("GET /orders/{id}", orderHandler.GetOrder)
	http.HandleFunc("PUT /orders/{id}", orderHandler.UpdateOrder)
	http.HandleFunc("DELETE /orders/{id}", orderHandler.DeleteOrder)
}
