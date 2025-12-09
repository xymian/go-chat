package main

import (
	"log"
	"net/http"

	"github.com/te6lim/go-chat/auth-service/routes"
	"github.com/te6lim/go-chat/auth-service/service"
)

func main() {
	routes.RegisterAuthRoutes()

	err := service.ConnectToUserService()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", service.Router)
	log.Println("Server started on localhost:50051")
	if err := http.ListenAndServe(":50051", nil); err != nil {
		log.Fatal(err)
	}
}
