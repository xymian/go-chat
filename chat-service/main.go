package main

import (
	"log"
	"net/http"
	"time"

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

	err := service.ConnectToUserService()
	if err != nil {
		log.Fatal(err)
	}

	routes.RegisterChatRoutes()

	go chat.ListenForActiveUsers()
	go chat.ListenForNewChatRoom()

	// periodically purge fully acknowledged messages that are not backed up
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			err := database.CleanupAllFullyAcknowledgedMessages()
			if err != nil {
				log.Println("cleanup error:", err)
			}
		}
	}()

	http.Handle("/", service.Router)
	log.Println("Server started on localhost:50053")
	if err := http.ListenAndServe(":50053", nil); err != nil {
		log.Fatal(err)
	}
}
