package database

import (
	"log"
)

type Message struct {
	Id               string `json:"id"`
	MessageReference string `json:"messageReference"`
	Text             string `json:"text"`
	SenderUsername   string `json:"senderUsername"`
	ReceiverUsername string `json:"receiver"`
	Timestamp        string `json:"timestamp"`
	ChatReference    string `json:"chatReference"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

func InsertMessage(message *Message) *Message {
	chat := GetChat(message.ChatReference)
	dbMessage := &Message{}
	err := Instance.QueryRow(
		`INSERT INTO messages (messageReference, text, senderUsername, receiverUsername, timestamp, chatReference)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, messageReference, text, senderUsername, receiverUsername, timestamp, chatReference, createdSt, updatedAt`,
		message.MessageReference, message.Text, message.SenderUsername, message.ReceiverUsername, message.Timestamp, chat.ChatReference,
	).Scan(
		&dbMessage.MessageReference, &dbMessage.Id, &dbMessage.Text, &dbMessage.SenderUsername, &dbMessage.ReceiverUsername, &dbMessage.Timestamp,
		&dbMessage.ChatReference, &dbMessage.CreatedAt, &dbMessage.UpdatedAt,
	)
	if err != nil {
		dbMessage = nil
	}
	return dbMessage
}

func GetMessage(chatReference string, messageReference string) *Message {
	newMessage := &Message{}
	err := Instance.QueryRow(
		`SELECT id, text, senderUsername, receiverUsername, timestamp, chatReference, createdAt, UpdatedAt FROM messages
		WHERE chatReference = $1 AND messageReference = $2`,
		chatReference, messageReference,
	).Scan(
		&newMessage.Id, &newMessage.MessageReference, &newMessage.Text, &newMessage.SenderUsername,
		&newMessage.ReceiverUsername, &newMessage.Timestamp, &newMessage.ChatReference, &newMessage.CreatedAt, &newMessage.UpdatedAt,
	)
	if err != nil {
		newMessage = nil
	}

	return newMessage
}

func GetAllMessages(chatReference string) []*Message {
	messages := []*Message{}
	rows, err := Instance.Query(
		`SELECT id, text, senderUsername, receiverUsername, timestamp, chatReference, createdAt, UpdatedAt FROM messages WHERE chatReference = $1`,
		chatReference,
	)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		message := &Message{}
		rows.Scan(
			&message.Id, &message.MessageReference, &message.Text, &message.SenderUsername, &message.ReceiverUsername, &message.Timestamp,
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
		RETURNING id, messageReference, text, senderUsername, receiverUsername, timestamp, chatReference, createdAt, updatedAt`,
	).Scan(
		&message.Id, &message.MessageReference, &message.Text, &message.SenderUsername, &message.ReceiverUsername,
		&message.Timestamp, &message.ChatReference, &message.CreatedAt, &message.UpdatedAt,
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
		RETURNING id, messageReference, text, senderUsername, receiverUsername, timestamp, chatReference, createdAt, updatedAt`,
		chatReference,
	)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		message := &Message{}
		rows.Scan(
			&message.Id, &message.MessageReference, &message.Text, &message.SenderUsername, &message.ReceiverUsername, &message.Timestamp,
			&message.ChatReference, &message.CreatedAt, &message.UpdatedAt,
		)
		messages = append(messages, message)
	}
	return messages
}
