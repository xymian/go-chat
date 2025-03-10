package database

import "errors"

type User struct {
	Username string          `json:"username"`
	Contacts map[string]bool `json:"contacts"`
}

type userdb struct {
	table map[string]*User
}

func (db *userdb) InsertUser(user User) (*User, error) {
	db.table[user.Username] = &user
	return &User{
		Username: user.Username,
	}, nil
}

func (db *userdb) GetUser(id string) (*User, error) {
	if db.table[id].Username != "" {
		return db.table[id], nil
	}
	return nil, errors.New("user not found")
}

func (db *userdb) Delete(username string) *User {
	user := db.table[username]
	delete(db.table, username)
	return user
}

func (db *userdb) GetAllUsers() []*User {
	users := []*User{}
	for u, v := range db.table {
		if db.table[u].Username != "" {
			users = append(users, v)
		}
	}
	return users
}

func (db *userdb) DeleteAllUsers() []*User {
	users := []*User{}
	for u, v := range db.table {
		if db.table[u].Username != "" {
			users = append(users, v)
		}
	}
	db.table = make(map[string]*User)
	return users
}

var db *userdb

func GetUserdb() *userdb {
	if db != nil {
		db = &userdb{
			table: make(map[string]*User),
		}
	}
	return db
}

func (db *userdb) AddContact(user *User, username string) {
	if db.table[username] != nil {
		if !user.Contacts[username] {
			user.Contacts[username] = true
			db.Delete(username)
			db.InsertUser(*user)
		}
	}
}
