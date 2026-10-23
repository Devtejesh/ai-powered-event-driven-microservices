package main

import (
	"fmt"
	"log"
	"net/http"

	"example.com/go-commerce-platform/user-service/config"
	"example.com/go-commerce-platform/user-service/handler"
	"example.com/go-commerce-platform/user-service/repository"
	"example.com/go-commerce-platform/user-service/route"
	"example.com/go-commerce-platform/user-service/service"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("user-service/.env"); err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	userrepo := repository.NewUserRepository(db)
	userservice := service.NewUserService(userrepo)
	authservice := service.NewAuthService(userrepo)
	authhandler := handler.NewAuthHandler(authservice)
	route.RegisterAuthRoutes(authhandler)
	userhandler := handler.NewUserHandler(userservice)

	route.RegisterUserRoutes(userhandler)

	fmt.Println("Server started running on port 8080")
	http.ListenAndServe(":8080", nil)

}
