package main

import (
	"log"
	"net/http"

	"github.com/te6lim/go-chat/chat-service/chat"
	"github.com/te6lim/go-chat/chat-service/database"
	"github.com/te6lim/go-chat/chat-service/routes"
	"github.com/te6lim/go-chat/chat-service/service"
)

func main() {
	database.ConnectToDB()
	pingErr := database.Instance.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	defer func() {
		database.Instance.Close()
		database.Instance = nil
	}()

	userService, err := service.ConnectToUserService()
	if err != nil {
		log.Fatal(err)
	}

	service.UserService = *userService

	routes.RegisterChatRoutes()

	go chat.ListenForActiveUsers()
	go chat.ListenForNewChatRoom()

	http.Handle("/", service.Router)
	log.Println("Server started on localhost:50053")
	if err := http.ListenAndServe(":50053", nil); err != nil {
		log.Fatal(err)
	}
}
