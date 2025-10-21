package main

import (
	"log"
	"net/http"

	"github.com/te6lim/go-chat/user-service/database"
	"github.com/te6lim/go-chat/user-service/routes"
	"github.com/te6lim/go-chat/user-service/service"
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

	routes.RegisterUserRoutes()

	http.Handle("/", service.Router)
	log.Println("Server started on localhost:5005")
	if err := http.ListenAndServe(":5005", nil); err != nil {
		log.Fatal(err)
	}
}
