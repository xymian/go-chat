package database

import (
	"errors"
	"log"
)

type Chat struct {
	Id            string `json:"id"`
	ChatReference string `json:"chatReference"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

func InsertChat(chat Chat) (*Chat, error) {
	newChat := &Chat{}
	if len(chat.ChatReference) == 0 {
		return nil, errors.New("invalid chat")
	}
	rows, err := Instance.Query(
		`INSERT INTO chats (chatReference) VALUES ($1) RETURNING id, chatReference, createdAt, updatedAt`,
		chat.ChatReference,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&chat.Id, &newChat.ChatReference, &newChat.CreatedAt, &newChat.UpdatedAt)
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
		`SELECT id, chatReference, createdAt, updatedAt FROM chats WHERE chatReference = $1`,
		reference,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&newChat.Id, &newChat.ChatReference, &newChat.CreatedAt, &newChat.UpdatedAt)
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
		`DELETE FROM chats WHERE chatReference = $1 RETURNING id, chatReference, createdAt, updatedAt`,
		reference,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&newChat.Id, &newChat.ChatReference, &newChat.CreatedAt, &newChat.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		return newChat, nil
	}

	return nil, nil
}
