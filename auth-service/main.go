package main

import (
	"log"
	"net/http"

	controllers "github.com/xymian/go-chat/auth-service/controllers"
	"github.com/xymian/go-chat/auth-service/routes"
	"github.com/xymian/go-chat/auth-service/service"
)

func main() {
	routes.RegisterAuthRoutes()

	var userService, err = service.ConnectToUserService()
	if err != nil {
		log.Fatal("unable to connect to user-service")
	}
	controllers.UserService = *userService
	
	http.Handle("/", service.Router)
	log.Println("Server started on localhost:50051")
	if err := http.ListenAndServe(":50051", nil); err != nil {
		log.Fatal(err)
	}
}
