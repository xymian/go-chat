package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type chatDB struct {
	messageTable map[string]*Message
	chatTable    map[string]*Chat
}

var mockdb *chatDB

var Instance *sql.DB

func ConnectToDB() {
	if Instance == nil {
		envErr := godotenv.Load()
		if envErr != nil {
			log.Fatal("Error loading .env file")
		}
		var dbUser = os.Getenv("DB_USER")
		var dbPassword = os.Getenv("DB_PASSWORD")
		var dbHost = os.Getenv("DB_HOST")
		var dbPort = os.Getenv("DB_PORT")
		var dbName = os.Getenv("DB_NAME")
		var ConnURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
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
			chatTable:    make(map[string]*Chat),
		}
	}
	return mockdb
}
