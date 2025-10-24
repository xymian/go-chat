package main

import (
	"log"
	"net/http"

	"github.com/xymian/go-chat/auth-service/routes"
	"github.com/xymian/go-chat/auth-service/service"
)

func main() {
	routes.RegisterAuthRoutes()

	var userService, err = service.ConnectToUserService()
	if err != nil {
		log.Fatal(err)
	}
	service.UserService = *userService
	
	http.Handle("/", service.Router)
	log.Println("Server started on localhost:50051")
	if err := http.ListenAndServe(":50051", nil); err != nil {
		log.Fatal(err)
	}
}
