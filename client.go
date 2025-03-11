package main

import (
	"fmt"
	"log"
	"net/http"

	chatserver "github.com/te6lim/go-chat/chat-server"
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/routes"
)

func main() {
	db := database.GetUserdb()

	db.InsertUser(database.User{
		Username: "teslim",
		Contacts: make(map[string]bool),
	})

	db.InsertUser(database.User{
		Username: "lateefah",
		Contacts: make(map[string]bool),
	})
	db.InsertUser(database.User{
		Username: "kikelomo",
		Contacts: make(map[string]bool),
	})
	db.InsertUser(database.User{
		Username: "mojo",
		Contacts: make(map[string]bool),
	})

	user := db.GetUser("teslim")
	db.AddContact(user, "lateefah")
	db.AddContact(user, "kikelomo")
	db.AddContact(user, "mojo")

	user = db.GetUser("lateefah")
	db.AddContact(user, "teslim")
	db.AddContact(user, "kikelomo")
	db.AddContact(user, "mojo")

	routes.RegisterUserRoutes(config.Router)
	routes.RegisterChatRoutes(config.Router)

	go chatserver.ListenForActiveUsers()
	go config.ListenForActiveSession()
	go config.ListenForCollectInputFlag()

	http.Handle("/", config.Router)
	log.Println("Server started on localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("server is running")
}
