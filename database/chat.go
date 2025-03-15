package database

type Chat struct {
	Id string
}

func (db *chatDB) InsertChat(chat Chat) *Chat {
	db.chatTable[chat.Id] = &chat
	return &Chat{
		Id: chat.Id,
	}
}

func (db *chatDB) GetChat(id string) *Chat {
	if db.chatTable[id] != nil {
		return db.chatTable[id]
	}
	return nil
}

func (db *chatDB) DeleteChat(id string) *Chat {
	chat := db.chatTable[id]
	delete(db.chatTable, id)
	return chat
}