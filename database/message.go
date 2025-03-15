package database

type Message struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Timestamp string `json:"timestamp"`
	ChatId    string `json:"chatId"`
	CreatedAt string `json:"createdAt"`
}

func (db *chatDB) InsertMessage(chatId string, message *Message) *Message {
	chat := db.GetChat(chatId)
	if chat == nil {
		chat = db.InsertChat(Chat{Id: chatId})
	}
	message.ChatId = chat.Id
	db.messageTable[message.Id] = message
	return message
}

func (db *chatDB) GetMessage(chatId string, messageId string) *Message {
	return db.messageTable[messageId]
}

func (db *chatDB) GetAllMessages(chatId string) []*Message {
	messages := []*Message{}
	for u, v := range db.messageTable {
		if db.messageTable[u].Id != "" {
			messages = append(messages, v)
		}
	}
	return messages
}

func (db *chatDB) DeleteMessage(chatId string, messageId string) *Message {
	message := db.messageTable[messageId]
	delete(db.messageTable, message.Id)
	return message
}

func (db *chatDB) DeleteAllMessages(chatId string) []*Message {
	messages := []*Message{}
	for u, v := range db.messageTable {
		if db.messageTable[u].Id != "" {
			messages = append(messages, v)
		}
	}
	db.messageTable = make(map[string]*Message)
	return messages
}
