package database

import (
	"errors"
	"log"
)

type ChatType string

const (
	ChatTypePrivate ChatType = "private"
	ChatTypeGroup   ChatType = "group"
)

type Chat struct {
	Id            string   `json:"id"`
	ChatReference string   `json:"chatReference"`
	ChatType      ChatType `json:"chatType"`
	Name          *string  `json:"name"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}

func InsertChat(chat Chat) (*Chat, error) {
	newChat := &Chat{}
	if len(chat.ChatReference) == 0 {
		return nil, errors.New("invalid chat")
	}
	if chat.ChatType == "" {
		chat.ChatType = ChatTypePrivate
	}
	rows, err := Instance.Query(
		`INSERT INTO chats (chatReference, chatType, name) VALUES ($1, $2, $3) RETURNING id, chatReference, chatType, name, createdAt, updatedAt`,
		chat.ChatReference, chat.ChatType, chat.Name,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&newChat.Id, &newChat.ChatReference, &newChat.ChatType, &newChat.Name, &newChat.CreatedAt, &newChat.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}

		return newChat, nil
	}
	return nil, nil
}

func GetChat(reference string) (*Chat, error) {
	newChat := &Chat{}
	rows, err := Instance.Query(
		`SELECT id, chatReference, chatType, name, createdAt, updatedAt FROM chats WHERE chatReference = $1`,
		reference,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&newChat.Id, &newChat.ChatReference, &newChat.ChatType, &newChat.Name, &newChat.CreatedAt, &newChat.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		return newChat, nil
	}
	return nil, nil
}

func DeleteChat(reference string) (*Chat, error) {
	newChat := &Chat{}
	rows, err := Instance.Query(
		`DELETE FROM chats WHERE chatReference = $1 RETURNING id, chatReference, chatType, name, createdAt, updatedAt`,
		reference,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&newChat.Id, &newChat.ChatReference, &newChat.ChatType, &newChat.Name, &newChat.CreatedAt, &newChat.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		return newChat, nil
	}

	return nil, nil
}
