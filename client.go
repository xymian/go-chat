package main

import (
	"fmt"
	"log"
	"net/http"

	chatserver "github.com/te6lim/go-chat/chat-server"
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/routes"
)

func main() {
	routes.RegisterUserRoutes(config.Router)
	routes.RegisterChatRoutes(config.Router)

	go chatserver.ListenForActiveUsers()
	go config.ListenForActiveSession()
	go config.ListenForCollectInputFlag()

	log.Println("Server started on localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("server is running")
}
