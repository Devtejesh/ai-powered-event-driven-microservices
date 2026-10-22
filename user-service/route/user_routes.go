package route

import (
	"net/http"

	"example.com/go-commerce-platform/user-service/handler"
)

func RegisterUserRoutes(userHandler *handler.UserHandler) {
	http.HandleFunc("POST /users", userHandler.CreateUser)
	http.HandleFunc("GET /users/{id}", userHandler.GetUserById)
	http.HandleFunc("DELETE /users/{id}", userHandler.DeleteUserBYId)
	http.HandleFunc("GET /users", userHandler.GetAllUsers)

}

func RegisterAuthRoutes(authHandler *handler.AuthHandler) {
	http.HandleFunc("POST /register", authHandler.RegisterUser)
	http.HandleFunc("POST /login", authHandler.LoginUser)
}
