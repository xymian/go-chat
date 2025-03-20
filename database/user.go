package database

import (
	"encoding/json"
	"log"
)

type User struct {
	Id        string          `json:"id"`
	Username  string          `json:"username"`
	Chats     map[string]bool `json:"chats"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
}

func CreateUsersTable() {
	_, err := Instance.Exec(
		`CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			chats JSON,
			createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)

	if err != nil {
		log.Fatalf("unable to create table: %v", err)
	}
}

func InsertUser(user User) *User {
	chatsJson, err := json.Marshal(&user.Chats)
	if err != nil {
		log.Fatal(err)
	}
	_, insertErr := Instance.Exec(
		`INSERT INTO users(username, chats) VALUES($1, $2)`, user.Username, string(chatsJson),
	)
	if insertErr != nil {
		log.Fatal(insertErr)
	}
	return &user
}

func GetUser(username string) *User {
	user := &User{}
	var chatsJson string = ""
	err := Instance.QueryRow(
		`SELECT id, username, chats, createdAt, updatedAt FROM users WHERE username = $1`, username,
	).Scan(&user.Id, &user.Username, &chatsJson, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		user = nil
	}
	jsonError := json.Unmarshal([]byte(chatsJson), &user.Chats)
	if jsonError != nil {
		log.Fatal(jsonError)
	}
	return user
}

func DeleteUser(username string) *User {
	user := &User{}
	var chatsJson string
	err := Instance.QueryRow(
		`DELETE FROM users WHERE username = $1 LIMIT 1 RETURNING id, username, chats, createdAt, updatedAt`, username,
	).Scan(&user.Id, &user.Username, &chatsJson, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Fatal(err)
	}
	jsonError := json.Unmarshal([]byte(chatsJson), &user.Chats)
	if jsonError != nil {
		log.Fatal(jsonError)
	}
	return user
}

func GetAllUsers() []*User {
	users := []*User{}
	rows, err := Instance.Query(
		`SELECT id, username, chats, createdAt, updatedAt FROM users`,
	)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		user := &User{}
		var chatsJson string
		err := rows.Scan(&user.Id, &user.Username, &chatsJson, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Fatal(err)
		}
		jsonError := json.Unmarshal([]byte(chatsJson), &user.Chats)
		if jsonError != nil {
			log.Fatal(jsonError)
		}
		users = append(users, user)
	}
	return users
}

func DeleteAllUsers() []*User {
	users := []*User{}
	rows, err := Instance.Query(
		`DELETE FROM users RETURNING id, username, chats, createdAt, updatedAt`,
	)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		user := &User{}
		var chatsJson string
		err := rows.Scan(&user.Id, &user.Username, &chatsJson, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Fatal(err)
		}
		jsonError := json.Unmarshal([]byte(chatsJson), &user.Chats)
		if jsonError != nil {
			log.Fatal(jsonError)
		}
		users = append(users, user)
	}
	return users
}

func UpdateUser(user User) *User {
	var chatsJson string
	updateErr := Instance.QueryRow(
		`UPDATE users SET username = $1, chats = $2 WHERE id = $3 RETURNING id, username, chats, createdAt, updatedAt`,
		user.Username, user.Chats, user.Id,
	).Scan(&user.Id, &user.Username, &chatsJson, &user.CreatedAt, &user.UpdatedAt)
	if updateErr != nil {
		log.Fatal(updateErr)
	}
	jsonError := json.Unmarshal([]byte(chatsJson), &user.Chats)
	if jsonError != nil {
		log.Fatal(jsonError)
	}
	return &user
}

func DropUsersTable() {
	_, err := Instance.Exec(`DROP TABLE users`)
	if err != nil {
		log.Fatal(err)
	}
}
