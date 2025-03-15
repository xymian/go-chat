package database

type chatDB struct {
	userTable map[string]*User
	messageTable map[string]*Message
	chatTable map[string]*Chat
	participantTable map[string]*Participant
}

var db *chatDB

func GetChatDB() *chatDB {
	if db == nil {
		db = &chatDB{
			messageTable: make(map[string]*Message),
			userTable: make(map[string]*User),
			chatTable: make(map[string]*Chat),
		}
	}
	return db
}