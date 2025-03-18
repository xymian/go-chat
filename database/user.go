package database

import (
	"encoding/json"
	"log"
)

type User struct {
	Id        string          `json:"id"`
	Username  string          `json:"username"`
	Contacts  map[string]bool `json:"contacts"`
	CreatedAt string          `json:"createdAt"`
}

func (db *chatDB) InsertUser(user User) *User {
	contactsJson, err := json.Marshal(user.Contacts)
	if err != nil {
		log.Fatal(err)
	}
	_, insertErr := Instance.Exec("INSERT INTO users(username, Contacts) VALUES($1, $2)", user.Username, string(contactsJson))
	if insertErr != nil {
		log.Fatal(err)
	}
	return &user
}

func (db *chatDB) GetUser(id string) *User {
	_, insertErr := Instance.Exec("GET * FROM users WHERE id = $1", id)
	if insertErr != nil {
		log.Fatal(insertErr)
	}
	if db.userTable[id] != nil {
		return db.userTable[id]
	}
	return nil
}

func (db *chatDB) Delete(username string) *User {
	user := db.userTable[username]
	delete(db.userTable, username)
	return user
}

func (db *chatDB) GetAllUsers() []*User {
	users := []*User{}
	for u, v := range db.userTable {
		if db.userTable[u].Username != "" {
			users = append(users, v)
		}
	}
	return users
}

func (db *chatDB) DeleteAllUsers() []*User {
	users := []*User{}
	for u, v := range db.userTable {
		if db.userTable[u].Username != "" {
			users = append(users, v)
		}
	}
	db.userTable = make(map[string]*User)
	return users
}

func (db *chatDB) AddContact(user *User, username string) {
	if db.userTable[username] != nil {
		if !user.Contacts[username] {
			user.Contacts[username] = true
			db.Delete(user.Username)
			db.InsertUser(*user)
		}
	}
}
