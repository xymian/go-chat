package database

import (
	"context"
	"encoding/json"
	"errors"
	"log"
)

type User struct {
	Id           int64   `json:"id"`
	Username     string  `json:"username"`
	PasswordHash string  `json:"passwordHash"`
	ChatReferences *string `json:"chatReferences"`
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
		rows, queryErr := txn.Query(
			`SELECT id, username FROM users WHERE id = $1`, id,
		)

		if queryErr != nil {
			log.Fatal(queryErr)
		}

		hasRows := rows.Next()
		if hasRows {
			scanErr := rows.Scan(&user.Id, &user.Username)
			if scanErr != nil {
				return nil, scanErr
			}
			users = append(users, *user)
		}
	}
	e := txn.Commit()
	if e != nil {
		return []User{}, e
	}
	return users, nil
}

func InsertUser(user User) (*User, error) {
	newUser := &User{}
	if len(user.Username) == 0 {
		return nil, errors.New("invalid username")
	}
	rows, err := Instance.Query(
		`INSERT INTO users(username, chatReferences, passwordHash) VALUES($1, $2, $3) RETURNING id, username, chatReferences, passwordHash, createdAt, updatedAt`,
		user.Username, user.ChatReferences, user.PasswordHash,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&newUser.Id, &newUser.Username, &user.ChatReferences, &newUser.PasswordHash, &newUser.CreatedAt, &newUser.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		return newUser, nil
	}
	return nil, nil
}

func GetUser(username string) (*User, error) {
	user := &User{}
	rows, err := Instance.Query(
		`SELECT id, username, passwordHash, ChatReferences, createdAt, updatedAt FROM users WHERE username = $1`, username,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.ChatReferences, &user.CreatedAt, &user.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		return user, nil
	}
	return nil, nil
}

func DeleteUser(username string) (*User, error) {
	user := &User{}
	rows, err := Instance.Query(
		`DELETE FROM users WHERE username = $1 LIMIT 1 RETURNING id, username, passwordHash, ChatReferences createdAt, updatedAt`, username,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.ChatReferences, &user.CreatedAt, &user.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		return user, nil
	}
	return nil, nil
}

func GetAllUsers() ([]User, error) {
	users := []User{}
	rows, err := Instance.Query(
		`SELECT id, username, passwordHash, ChatReferences, createdAt, updatedAt FROM users`,
	)
	if err != nil {
		return []User{}, nil
	}

	for rows.Next() {
		user := User{}
		scanErr := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.ChatReferences, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, scanErr
		}
		users = append(users, user)
	}

	defer rows.Close()

	return users, nil
}

func DeleteAllUsers() ([]User, error) {
	users := []User{}
	rows, err := Instance.Query(
		`DELETE FROM users RETURNING id, username, passwordHash, ChatReferences, createdAt, updatedAt`,
	)
	if err != nil {
		return []User{}, nil
	}

	defer rows.Close()

	for rows.Next() {
		user := User{}
		scanErr := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.ChatReferences, &user.CreatedAt, &user.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		users = append(users, user)
	}
	return users, nil
}

func AddConversation(user *User, otherUserId int64, chatReference string) (*User, error) {
	newUser := &User{}
	conversationMap := map[int64]string{}
	if user.ChatReferences != nil {
		json.Unmarshal([]byte(*user.ChatReferences), &conversationMap)
	}
	conversationMap[otherUserId] = chatReference
	ChatReferencesJsonString, err := json.Marshal(conversationMap)
	if err != nil {
		return nil, err
	}
	rows, queryErr := Instance.Query(
		`UPDATE users SET ChatReferences = $1 WHERE id = $2 RETURNING id, username, passwordHash, ChatReferences, createdAt, updatedAt`,
		ChatReferencesJsonString, user.Id,
	)

	if queryErr != nil {
		log.Fatal(queryErr)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&newUser.Id, &newUser.Username, &newUser.PasswordHash, &newUser.ChatReferences, &newUser.CreatedAt, &newUser.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}

		return newUser, nil
	}
	return nil, nil
}
