package database

import (
	"encoding/json"
	"errors"
	"log"
)

type User struct {
	Id           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	Conversations *string `json:"conversations"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

func InsertUser(user User) (*User, error) {
	newUser := &User{}
	if len(user.Username) <= 0 {
		return nil, errors.New("invalid username")
	}
	err := Instance.QueryRow(
		`INSERT INTO users(username, passwordHash) VALUES($1, $2) RETURNING id, username, passwordHash, createdAt, updatedAt`,
		user.Username, user.PasswordHash,
	).Scan(&newUser.Id, &newUser.Username, &newUser.PasswordHash, &newUser.CreatedAt, &newUser.UpdatedAt)
	if err != nil {
		newUser = nil
	}
	return newUser, err
}

func GetUser(username string) *User {
	user := &User{}
	err := Instance.QueryRow(
		`SELECT id, username, passwordHash, conversations, createdAt, updatedAt FROM users WHERE username = $1`, username,
	).Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Conversations, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		user = nil
		println("nil user")
	}

	return user
}

func DeleteUser(username string) *User {
	user := &User{}
	err := Instance.QueryRow(
		`DELETE FROM users WHERE username = $1 LIMIT 1 RETURNING id, username, passwordHash, conversations createdAt, updatedAt`, username,
	).Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Conversations, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		user = nil
	}
	return user
}

func GetAllUsers() []*User {
	users := []*User{}
	rows, err := Instance.Query(
		`SELECT id, username, passwordHash, conversations, createdAt, updatedAt FROM users`,
	)
	if err != nil {
		users = nil
	}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Conversations, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return users
}

func DeleteAllUsers() []*User {
	users := []*User{}
	rows, err := Instance.Query(
		`DELETE FROM users RETURNING id, username, passwordHash, conversations, createdAt, updatedAt`,
	)
	if err != nil {
		users = nil
	}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Conversations, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return users
}

func (user *User) AddConversation(username string, chatReference string) (*User, error) {
	newUser := &User{}
	conversationMap := map[string]string {}
	if (user.Conversations != nil) {
		json.Unmarshal([]byte(*user.Conversations), &conversationMap)
	}
	conversationMap[username] = chatReference
	conversationsJsonString, err := json.Marshal(conversationMap)
	if err != nil {
		return nil, errors.New("unable to add new conversation")
	}
	updateErr := Instance.QueryRow(
		`UPDATE users SET conversations = $1 WHERE id = $2 RETURNING id, username, passwordHash, conversations, createdAt, updatedAt`,
		conversationsJsonString, user.Id,
	).Scan(&newUser, &newUser.Username, &newUser.PasswordHash, &newUser.Conversations, &newUser.CreatedAt, &newUser.UpdatedAt)
	if updateErr != nil {
		newUser = nil
	}
	return newUser, nil
}

/*func UpdateUser(user User) (*User, error) {
	newUser := &User{}
	updateErr := Instance.QueryRow(
		`UPDATE users SET username = $1 WHERE id = $3 RETURNING id, username, createdAt, updatedAt`,
		user.Username, user.Id,
	).Scan(&newUser, &newUser.Username, &newUser.CreatedAt, &newUser.UpdatedAt)
	if updateErr != nil {
		newUser = nil
	}
	return newUser, nil
}*/
