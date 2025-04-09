package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5"
	"github.com/te6lim/go-chat/config"
)

var Instance *sql.DB

func ConnectToDB() {
	if Instance == nil {
		if config.EnvErr != nil {
			log.Fatal("Error loading .env file")
		}
		var dbUser = os.Getenv("DB_USER")
		var dbPassword = os.Getenv("DB_PASSWORD")
		var dbHost = os.Getenv("DB_HOST")
		var dbPort = os.Getenv("DB_PORT")
		var dbName = os.Getenv("DB_NAME")
		var sslmode = os.Getenv("SSL_MODE")
		var connURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPassword, dbHost, dbPort, dbName, sslmode)
		newdb, err := sql.Open("pgx", connURL)
		if err != nil {
			log.Fatal(err)
		}
		Instance = newdb

		mig, migErr := migrate.New("file://database/migrations", connURL)
		if migErr != nil {
			log.Fatal("error creating migration instance: ", migErr)
		}
		if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("error apllying migrations: ", err)
		}
		fmt.Println("successfully migrated db")
	}
}
