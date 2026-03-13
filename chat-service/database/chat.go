package database

import (
	"errors"
	"log"
)

type Chat struct {
	Id            string  `json:"id"`
	ChatReference string  `json:"chatReference"`
	IsGroup       bool    `json:"isGroup"`
	Name          *string `json:"name"`
	CreatedBy     *string `json:"createdBy"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

func InsertChat(chat Chat) (*Chat, error) {
	newChat := &Chat{}
	if len(chat.ChatReference) == 0 {
		return nil, errors.New("invalid chat")
	}
	rows, err := Instance.Query(
		`INSERT INTO chats (chatReference, is_group, name, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, chatReference, is_group, name, created_by, createdAt, updatedAt`,
		chat.ChatReference, chat.IsGroup, chat.Name, chat.CreatedBy,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&newChat.Id, &newChat.ChatReference, &newChat.IsGroup,
			&newChat.Name, &newChat.CreatedBy, &newChat.CreatedAt, &newChat.UpdatedAt,
		)
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
		`SELECT id, chatReference, is_group, name, created_by, createdAt, updatedAt
		FROM chats WHERE chatReference = $1`,
		reference,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&newChat.Id, &newChat.ChatReference, &newChat.IsGroup,
			&newChat.Name, &newChat.CreatedBy, &newChat.CreatedAt, &newChat.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return newChat, nil
	}
	return nil, nil
}

func IsGroupChat(chatReference string) (bool, error) {
	chat, err := GetChat(chatReference)
	if err != nil {
		return false, err
	}
	if chat == nil {
		return false, errors.New("chat not found")
	}
	return chat.IsGroup, nil
}

func UpdateGroupName(chatReference string, name string) (*Chat, error) {
	newChat := &Chat{}
	rows, err := Instance.Query(
		`UPDATE chats SET name = $1, updatedAt = NOW()
		WHERE chatReference = $2 AND is_group = TRUE
		RETURNING id, chatReference, is_group, name, created_by, createdAt, updatedAt`,
		name, chatReference,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&newChat.Id, &newChat.ChatReference, &newChat.IsGroup,
			&newChat.Name, &newChat.CreatedBy, &newChat.CreatedAt, &newChat.UpdatedAt,
		)
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
		`DELETE FROM chats WHERE chatReference = $1
		RETURNING id, chatReference, is_group, name, created_by, createdAt, updatedAt`,
		reference,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&newChat.Id, &newChat.ChatReference, &newChat.IsGroup,
			&newChat.Name, &newChat.CreatedBy, &newChat.CreatedAt, &newChat.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return newChat, nil
	}
	return nil, nil
}
