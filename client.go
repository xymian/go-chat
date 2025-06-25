package main

import (
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/te6lim/go-chat/chat"
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/routes"
)

func main() {

	database.ConnectToDB()

	pingErr := database.Instance.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	defer database.Instance.Close()

	routes.RegisterUserRoutes()
	routes.RegisterChatRoutes()
	routes.RegisterAuthRoutes()

	go chat.ListenForActiveUsers()
	go chat.ListenForNewChatRoom()

	http.Handle("/", config.Router)
	log.Println("Server started on localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
