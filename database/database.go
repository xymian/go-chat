package database

import (
	"database/sql"
	"fmt"
	"log"
)

type chatDB struct {
	userTable    map[string]*User
	messageTable map[string]*Message
	chatTable    map[string]*Chat
}

var mockdb *chatDB

var Instance *sql.DB

var ConnURL = "postgres://postgres:plantainchips@localhost:5432/go-chat"

func ConnectToDB() {
	if Instance == nil {
		newdb, err := sql.Open("pgx", ConnURL)
		if err != nil {
			log.Fatal()
		}
		Instance = newdb
		fmt.Println("successfully connected to go-chat database")
	}
}

func GetChatDB() *chatDB {
	if mockdb == nil {
		mockdb = &chatDB{
			messageTable: make(map[string]*Message),
			userTable:    make(map[string]*User),
			chatTable:    make(map[string]*Chat),
		}
	}
	return mockdb
}
