package database

import (
	"encoding/json"
	"log"
)

type Chat struct {
	Id            string          `json:"id"`
	ChatReference string          `json:"chatReference"`
	Participants  map[string]bool `json:"participants"`
	CreatedAt     string          `json:"createdAt"`
	UpdatedAt     string          `json:"updatedAt"`
}

func CreateChatsTable() {
	_, err := Instance.Exec(
		`CREATE TABLE chats (
			id SERIAL PRIMARY KEY,
			chatReference TEXT NOT NUll,
			participants JSON
			createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	)
	if err != nil {
		log.Fatalf("unable to create table: %v", err)
	}
}

func InsertChat(chat Chat) *Chat {
	newChat := &Chat{}
	var participants string
	err := Instance.QueryRow(
		`INSERT INTO chats (chatReference, participants) VALUES ($1, $2) RETURNING id, chatReference, participants, createdAt, updatedAt`,
		chat.ChatReference, chat.Participants,
	).Scan(&chat.Id, &newChat.ChatReference, &participants, &newChat.CreatedAt, &newChat.UpdatedAt)
	if err != nil {
		log.Fatal(err)
	}
	jsonerr := json.Unmarshal([]byte(participants), &newChat.Participants)
	if jsonerr != nil {
		log.Fatal(jsonerr)
	}
	return newChat
}

func GetChat(reference string) *Chat {
	newChat := &Chat{}
	var participants string = ""
	err := Instance.QueryRow(
		`SELECT Id, ChatReference, Participants, CreatedAt, UpdatedAt FRON chats WHERE chatReference = $1`,
		reference,
	).Scan(&newChat.Id, &newChat.ChatReference, &participants, &newChat.CreatedAt, &newChat.UpdatedAt)
	if err != nil {
		newChat = nil
	}
	jsonerr := json.Unmarshal([]byte(participants), &newChat.Participants)
	if jsonerr != nil {
		log.Fatal(jsonerr)
	}
	return newChat
}

func DeleteChat(reference string) *Chat {
	newChat := &Chat{}
	var participants string = ""
	err := Instance.QueryRow(
		`DELETE FROM chats WHERE chatReference = $1 RETURNING id, chatReference, participants, createdAt, updatedAt`,
		reference,
	).Scan(&newChat.Id, &newChat.ChatReference, &participants, &newChat.CreatedAt, &newChat.UpdatedAt)
	if err != nil {
		newChat = nil
	}
	jsonerr := json.Unmarshal([]byte(participants), &newChat.Participants)
	if jsonerr != nil {
		log.Fatal(jsonerr)
	}
	return newChat
}

func DropChatsTable() {
	_, err := Instance.Exec(`DROP TABLE chats`)
	if err != nil {
		log.Fatal(err)
	}
}
