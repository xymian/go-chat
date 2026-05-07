package database

import (
	"context"
	"errors"
	"log"
)

type User struct {
	Id           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	DisplayName  string `json:"displayName"`
	Bio          string `json:"bio"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
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
		`INSERT INTO users(username, passwordHash) VALUES($1, $2)
		RETURNING id, username, passwordHash, displayName, bio, createdAt, updatedAt`,
		user.Username, user.PasswordHash,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&newUser.Id, &newUser.Username, &newUser.PasswordHash, &newUser.DisplayName, &newUser.Bio, &newUser.CreatedAt, &newUser.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		return newUser, nil
	}
	return nil, nil
}

func UpdateUserProfile(username string, displayName string, bio string) (*User, error) {
	user := &User{}
	rows, err := Instance.Query(
		`UPDATE users SET displayName = $1, bio = $2, updatedAt = CURRENT_TIMESTAMP
		WHERE username = $3
		RETURNING id, username, passwordHash, displayName, bio, createdAt, updatedAt`,
		displayName, bio, username,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.DisplayName, &user.Bio, &user.CreatedAt, &user.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		return user, nil
	}
	return nil, nil
}

func GetUser(username string) (*User, error) {
	user := &User{}
	rows, err := Instance.Query(
		`SELECT id, username, passwordHash, displayName, bio, createdAt, updatedAt FROM users WHERE username = $1`, username,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.DisplayName, &user.Bio, &user.CreatedAt, &user.UpdatedAt)
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
		`DELETE FROM users WHERE username = $1 RETURNING id, username, passwordHash, displayName, bio, createdAt, updatedAt`, username,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.DisplayName, &user.Bio, &user.CreatedAt, &user.UpdatedAt)
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
		`SELECT id, username, passwordHash, displayName, bio, createdAt, updatedAt FROM users`,
	)
	if err != nil {
		return []User{}, nil
	}

	defer rows.Close()

	for rows.Next() {
		user := User{}
		scanErr := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.DisplayName, &user.Bio, &user.CreatedAt, &user.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		users = append(users, user)
	}

	return users, nil
}

func DeleteAllUsers() ([]User, error) {
	users := []User{}
	rows, err := Instance.Query(
		`DELETE FROM users RETURNING id, username, passwordHash, displayName, bio, createdAt, updatedAt`,
	)
	if err != nil {
		return []User{}, nil
	}

	defer rows.Close()

	for rows.Next() {
		user := User{}
		scanErr := rows.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.DisplayName, &user.Bio, &user.CreatedAt, &user.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		users = append(users, user)
	}
	return users, nil
}
