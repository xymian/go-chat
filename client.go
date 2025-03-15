package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/te6lim/go-chat/chat"
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/routes"
)

func main() {
	db := database.GetChatDB()

	db.InsertUser(database.User{
		Username: "user0",
		Contacts: make(map[string]bool),
	})

	db.InsertUser(database.User{
		Username: "user1",
		Contacts: make(map[string]bool),
	})
	db.InsertUser(database.User{
		Username: "user2",
		Contacts: make(map[string]bool),
	})
	db.InsertUser(database.User{
		Username: "user3",
		Contacts: make(map[string]bool),
	})

	user := db.GetUser("user0")
	db.AddContact(user, "user1")
	db.AddContact(user, "user2")
	db.AddContact(user, "user3")

	user = db.GetUser("user1")
	db.AddContact(user, "user0")
	db.AddContact(user, "user2")
	db.AddContact(user, "user3")

	routes.RegisterUserRoutes(config.Router)
	routes.RegisterChatRoutes(config.Router)

	go chat.ListenForActiveUsers()
	go config.ListenForCollectInputFlag()

	http.Handle("/", config.Router)
	log.Println("Server started on localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("server is running")
}
