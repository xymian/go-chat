package database

import (
	"log"
)

type User struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func InsertUser(user User) *User {
	newUser := &User{}
	err := Instance.QueryRow(
		`INSERT INTO users(username) VALUES($1)`,
		user.Username,
	).Scan(&newUser.Id, &newUser.Username, &newUser.CreatedAt, &newUser.UpdatedAt)
	if err != nil {
		newUser = nil
	}
	return newUser
}

func GetUser(username string) *User {
	user := &User{}
	err := Instance.QueryRow(
		`SELECT id, username, createdAt, updatedAt FROM users WHERE username = $1`, username,
	).Scan(&user.Id, &user.Username, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		user = nil
	}

	return user
}

func DeleteUser(username string) *User {
	user := &User{}
	err := Instance.QueryRow(
		`DELETE FROM users WHERE username = $1 LIMIT 1 RETURNING id, username, createdAt, updatedAt`, username,
	).Scan(&user.Id, &user.Username, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		user = nil
	}
	return user
}

func GetAllUsers() []*User {
	users := []*User{}
	rows, err := Instance.Query(
		`SELECT id, username, createdAt, updatedAt FROM users`,
	)
	if err != nil {
		users = nil
	}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.Username, &user.CreatedAt, &user.UpdatedAt)
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
		`DELETE FROM users RETURNING id, username, createdAt, updatedAt`,
	)
	if err != nil {
		users = nil
	}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.Id, &user.Username, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}
	return users
}

func UpdateUser(user User) *User {
	newUser := &User{}
	updateErr := Instance.QueryRow(
		`UPDATE users SET username = $1 WHERE id = $3 RETURNING id, username, createdAt, updatedAt`,
		user.Username, user.Id,
	).Scan(&newUser, &newUser.Username, &newUser.CreatedAt, &newUser.UpdatedAt)
	if updateErr != nil {
		newUser = nil
	}
	return newUser
}
