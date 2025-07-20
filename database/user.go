package database

import (
	"context"
	"encoding/json"
	"errors"
)

type User struct {
	Id           int64   `json:"id"`
	Username     string  `json:"username"`
	PasswordHash string  `json:"passwordHash"`
	Interactions *string `json:"interactions"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

func GetUsers(ids ...int64) ([]User, error) {
	ctx := context.Background()
	txn, err := Instance.BeginTx(ctx, nil)
	if err != nil {
		return []User{}, err
	}
	defer txn.Rollback()

	users := []User{}
	for _, id := range ids {
		user := &User{}
		err := txn.QueryRow(
			`SELECT id, username FROM users WHERE id = $1`, id,
		).Scan(&user.Id, &user.Username)
		if err != nil {
			return []User{}, err
		}
		users = append(users, *user)
	}
	e := txn.Commit()
	if e != nil {
		return []User{}, err
	}
	return users, nil
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
		return nil, err
	}
	return newUser, nil
}

func GetUser(username string) (*User, error) {
	user := &User{}
	err := Instance.QueryRow(
		`SELECT id, username, passwordHash, interactions, createdAt, updatedAt FROM users WHERE username = $1`, username,
	).Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Interactions, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func DeleteUser(username string) (*User, error) {
	user := &User{}
	err := Instance.QueryRow(
		`DELETE FROM users WHERE username = $1 LIMIT 1 RETURNING id, username, passwordHash, interactions createdAt, updatedAt`, username,
	).Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Interactions, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetAllUsers() ([]*User, error) {
	users := []*User{}
	rows, err := Instance.Query(
		`SELECT id, username, passwordHash, interactions, createdAt, updatedAt FROM users`,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Interactions, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func DeleteAllUsers() ([]*User, error) {
	users := []*User{}
	rows, err := Instance.Query(
		`DELETE FROM users RETURNING id, username, passwordHash, interactions, createdAt, updatedAt`,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Interactions, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (user *User) AddConversation(userId int64, chatReference string) (*User, error) {
	newUser := &User{}
	conversationMap := map[int64]string{}
	if user.Interactions != nil {
		json.Unmarshal([]byte(*user.Interactions), &conversationMap)
	}
	conversationMap[userId] = chatReference
	interactionsJsonString, err := json.Marshal(conversationMap)
	if err != nil {
		return nil, err
	}
	Instance.QueryRow(
		`UPDATE users SET interactions = $1 WHERE id = $2 RETURNING id, username, passwordHash, interactions, createdAt, updatedAt`,
		interactionsJsonString, user.Id,
	).Scan(&newUser, &newUser.Username, &newUser.PasswordHash, &newUser.Interactions, &newUser.CreatedAt, &newUser.UpdatedAt)
	
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
