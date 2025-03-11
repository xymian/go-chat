package database


type User struct {
	Username string          `json:"username"`
	Contacts map[string]bool `json:"contacts"`
}

type userdb struct {
	table map[string]*User
}

func (db *userdb) InsertUser(user User) *User {
	db.table[user.Username] = &user
	return &User{
		Username: user.Username,
		Contacts: user.Contacts,
	}
}

func (db *userdb) GetUser(id string) *User {
	if db.table[id] != nil {
		return db.table[id]
	}
	return nil
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
	if db == nil {
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
			db.Delete(user.Username)
			db.InsertUser(*user)
		}
	}
}
