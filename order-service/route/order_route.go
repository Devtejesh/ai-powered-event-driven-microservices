package route

import (
	"net/http"

	"example.com/go-commerce-platform/order-service/handler"
	"example.com/go-commerce-platform/order-service/middleware"
)

func RegisterOrderRoutes(orderHandler *handler.OrderHandler) {
	http.Handle("POST /orders",
		middleware.AuthMiddleware(

			http.HandlerFunc(orderHandler.CreateOrder),
		),
	)

	http.Handle("GET /orders",
		middleware.AuthMiddleware(
			middleware.RequireAdmin(
				http.HandlerFunc(orderHandler.ListOrders),
			),
		),
	)

	http.Handle("GET /orders/{id}",
		middleware.AuthMiddleware(
			middleware.RequireAdmin(
				http.HandlerFunc(orderHandler.GetOrder),
			),
		),
	)

	http.Handle("PUT /orders/{id}",
		middleware.AuthMiddleware(
			middleware.RequireAdmin(
				http.HandlerFunc(orderHandler.UpdateOrder),
			),
		),
	)

	http.Handle("DELETE /orders/{id}",
		middleware.AuthMiddleware(
			middleware.RequireAdmin(
				http.HandlerFunc(orderHandler.DeleteOrder),
			),
		),
	)
}
