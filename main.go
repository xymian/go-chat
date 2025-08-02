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

	defer func(){
		database.Instance.Close()
		database.Instance = nil
	}()

	routes.RegisterUserRoutes()
	routes.RegisterChatRoutes()
	routes.RegisterAuthRoutes()

	go chat.ListenForActiveUsers()
	go chat.ListenForNewChatRoom()

	http.Handle("/", config.Router)
	log.Println("Server started on localhost:6060")
	if err := http.ListenAndServe(":6060", nil); err != nil {
		log.Fatal(err)
	}
}
