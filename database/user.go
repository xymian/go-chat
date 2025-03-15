package database

type User struct {
	Id        string          `json:"id"`
	Username  string          `json:"username"`
	Contacts  map[string]bool `json:"contacts"`
	CreatedAt string          `json:"createdAt"`
}

func (db *chatDB) InsertUser(user User) *User {
	db.userTable[user.Username] = &user
	return &User{
		Username: user.Username,
		Contacts: user.Contacts,
	}
}

func (db *chatDB) GetUser(id string) *User {
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
