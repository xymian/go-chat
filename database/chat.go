package database

import (
	"log"
)

type Chat struct {
	Id            string `json:"id"`
	ChatReference string `json:"chatReference"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

func InsertChat(chat Chat) *Chat {
	newChat := &Chat{}
	err := Instance.QueryRow(
		`INSERT INTO chats(chatreference) VALUES ($1) RETURNING id, chatreference, createdAt, updatedAt`,
		chat.ChatReference,
	).Scan(&chat.Id, &newChat.ChatReference, &newChat.CreatedAt, &newChat.UpdatedAt)
	if err != nil {
		log.Fatal(err)
	}
	return newChat
}

func GetChat(reference string) *Chat {
	newChat := &Chat{}
	err := Instance.QueryRow(
		`SELECT id, chatReference, createdAt, updatedAt FROM chats WHERE chatReference = $1`,
		reference,
	).Scan(&newChat.Id, &newChat.ChatReference, &newChat.CreatedAt, &newChat.UpdatedAt)
	if err != nil {
		newChat = nil
	}
	return newChat
}

func DeleteChat(reference string) *Chat {
	newChat := &Chat{}
	err := Instance.QueryRow(
		`DELETE FROM chats WHERE chatReference = $1 RETURNING id, chatReference, createdAt, updatedAt`,
		reference,
	).Scan(&newChat.Id, &newChat.ChatReference, &newChat.CreatedAt, &newChat.UpdatedAt)
	if err != nil {
		newChat = nil
	}
	return newChat
}

func DropChatsTable() {
	_, err := Instance.Exec(`DROP TABLE chats`)
	if err != nil {
		log.Fatal(err)
	}
}
