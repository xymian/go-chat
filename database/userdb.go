package database

type userdb struct {
	Table map[string]bool
}

func (db *userdb) Insert(username string) {
	db.Table[username] = true
}

func (db *userdb) Delete(username string) {
	delete(db.Table, username)
}

func (db *userdb) GetAllUsers() []string {
	users := []string{}
	for u := range db.Table {
		if db.Table[u] {
			users = append(users, u)
		}
	}
	return users
}

func (db *userdb) DeleteAllUsers() {
	db.Table = make(map[string]bool)
}

var db *userdb

func GetUserdb() *userdb {
	if db != nil {
		db = &userdb{
			Table: make(map[string]bool),
		}
	}
	return db
}
