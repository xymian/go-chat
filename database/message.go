package database

import (
	"errors"
	"log"
)

type Message struct {
	Id               string `json:"id"`
	MessageReference string `json:"messageReference"`
	TextMessage      string `json:"textMessage"`
	SenderUsername   string `json:"senderUsername"`
	ReceiverUsername string `json:"receiverUsername"`
	MessageTimestamp string `json:"messageTimestamp"`
	ChatReference    string `json:"chatReference"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

func InsertMessage(message Message) (*Message, error) {
	chat := GetChat(message.ChatReference)
	if chat == nil {
		return nil, errors.New("chat reference for this message does not exits")
	}
	var messageErr error = nil
	switch {
	case message.MessageReference == "":
		messageErr = errors.New("message reference cannot be empty")
	case message.TextMessage == "":
		messageErr = errors.New("message text cannot be empty")
	case message.SenderUsername == "":
		messageErr = errors.New("message sender cannot be empty")
	case message.ReceiverUsername == "":
		messageErr = errors.New("message receiver cannot be empty")
	case message.MessageTimestamp == "":
		messageErr = errors.New("message timestamp cannot be empty")
	}
	if messageErr != nil {
		return nil, messageErr
	}
	dbMessage := &Message{}
	err := Instance.QueryRow(
		`INSERT INTO messages (messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp, chatReference)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp, chatReference, createdAt, updatedAt`,
		message.MessageReference, message.TextMessage, message.SenderUsername, message.ReceiverUsername, message.MessageTimestamp, chat.ChatReference,
	).Scan(
		&dbMessage.Id, &dbMessage.MessageReference, &dbMessage.TextMessage, &dbMessage.SenderUsername, &dbMessage.ReceiverUsername, &dbMessage.MessageTimestamp,
		&dbMessage.ChatReference, &dbMessage.CreatedAt, &dbMessage.UpdatedAt,
	)
	if err != nil {
		dbMessage = nil
		messageErr = err
	}
	return dbMessage, messageErr
}

func GetMessage(chatReference string, messageReference string) *Message {
	newMessage := &Message{}
	err := Instance.QueryRow(
		`SELECT id, textMessage, senderUsername, receiverUsername, messageTimestamp, chatReference, createdAt, UpdatedAt FROM messages
		WHERE chatReference = $1 AND messageReference = $2`,
		chatReference, messageReference,
	).Scan(
		&newMessage.Id, &newMessage.MessageReference, &newMessage.TextMessage, &newMessage.SenderUsername,
		&newMessage.ReceiverUsername, &newMessage.MessageTimestamp, &newMessage.ChatReference, &newMessage.CreatedAt, &newMessage.UpdatedAt,
	)
	if err != nil {
		newMessage = nil
	}

	return newMessage
}

func GetAllMessages(chatReference string) []*Message {
	messages := []*Message{}
	rows, err := Instance.Query(
		`SELECT id, textMessage, senderUsername, receiverUsername, messageTimestamp, chatReference, createdAt, UpdatedAt FROM messages WHERE chatReference = $1`,
		chatReference,
	)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		message := &Message{}
		rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername, &message.ReceiverUsername, &message.MessageTimestamp,
			&message.ChatReference, &message.CreatedAt, &message.UpdatedAt,
		)
		messages = append(messages, message)
	}
	return messages
}

func DeleteMessage(messageReference string) *Message {
	message := &Message{}
	err := Instance.QueryRow(
		`DELETE FROM messages WHERE messageReference = $1
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp, chatReference, createdAt, updatedAt`,
	).Scan(
		&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername, &message.ReceiverUsername,
		&message.MessageTimestamp, &message.ChatReference, &message.CreatedAt, &message.UpdatedAt,
	)
	if err != nil {
		message = nil
	}
	return message
}

func DeleteAllMessages(chatReference string) []*Message {
	messages := []*Message{}
	rows, err := Instance.Query(
		`DELETE FROM messages WHERE chatReference = $1,
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp, chatReference, createdAt, updatedAt`,
		chatReference,
	)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		message := &Message{}
		rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername, &message.ReceiverUsername, &message.MessageTimestamp,
			&message.ChatReference, &message.CreatedAt, &message.UpdatedAt,
		)
		messages = append(messages, message)
	}
	return messages
}
