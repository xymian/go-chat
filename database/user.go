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

func InsertUser(user User) *User {
	_, insertErr := Instance.Exec("INSERT INTO users(username, chats) VALUES($1, $2)", user.Username, user.Chats)
	if insertErr != nil {
		log.Fatal(insertErr)
	}
	return &user
}

func GetUser(username string) *User {
	user := &User{}
	var chatsJson string
	err := Instance.QueryRow(
		"SELECT id, username, contacts, createdAt, updatedAt FROM users WHERE username = $1", username,
	).Scan(&user.Id, &user.Username, chatsJson, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Fatal(err)
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
		"DELETE FROM users WHERE username = $1 LIMIT 1 RETURNING id, username, chats, createdAt, updatedAt", username,
	).Scan(&user.Id, &user.Username, chatsJson, &user.CreatedAt, &user.UpdatedAt)
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
	rows, err := Instance.Query("SELECT * FROM users")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		user := &User{}
		var chatsJson string
		err := rows.Scan(&user.Id, &user.Username, chatsJson, &user.CreatedAt, &user.UpdatedAt)
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
	rows, err := Instance.Query("DELETE FROM users RETURNING id, username, chats, createdAt, updatedAt")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		user := &User{}
		var chatsJson string
		err := rows.Scan(&user.Id, &user.Username, chatsJson, &user.CreatedAt, &user.UpdatedAt)
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
	updateErr := Instance.QueryRow(
		"UPDATE users SET username = $1, chats = $2 WHERE id = $3 RETURNING id, username, chats, createdAt, updatedAt",
		user.Username, user.Chats, user.Id,
	).Scan(&user.Id, &user.Username, &user.Chats, &user.CreatedAt, &user.UpdatedAt)
	if updateErr != nil {
		log.Fatal(updateErr)
	}
	return &user
}
