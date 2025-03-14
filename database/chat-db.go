package database

type Message struct {
	Id string
	Text string
	Sender string
	Receiver string
	Timestamp string
}

type chatDB struct {
	table map[string]*Message
}

var chatDBs = make(map[string]*chatDB)

func (db *chatDB) InsertMessage(message *Message) *Message {
	db.table[message.Id] = message
	return message
}

func (db *chatDB) GetMessage(id string) *Message {
	return db.table[id]
}

func (db *chatDB) GetAllMessages() []*Message {
	messages := []*Message{}
	for u, v := range db.table {
		if db.table[u].Id != "" {
			messages = append(messages, v)
		}
	}
	return messages
}

func (db *chatDB) DeleteMessage(id string) *Message {
	message := db.table[id]
	delete(db.table, message.Id)
	return message
}

func (db *chatDB) DeleteAllMessages() []*Message {
	messages := []*Message{}
	for u, v := range db.table {
		if db.table[u].Id != "" {
			messages = append(messages, v)
		}
	}
	db.table = make(map[string]*Message)
	return messages
}

func GetChatDB(sharedUserId string) *chatDB {
	chatdb := chatDBs[sharedUserId]
	if chatdb == nil {
		chatdb = &chatDB{
			table: make(map[string]*Message),
		}
		chatDBs[sharedUserId] = chatdb
	}
	return chatdb
}