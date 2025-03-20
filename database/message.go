package database

import (
	"log"
)

type Message struct {
	Id               string `json:"id"`
	MessageReference string `json:"messageReference"`
	Text             string `json:"text"`
	Sender           string `json:"sender"`
	Receiver         string `json:"receiver"`
	Timestamp        string `json:"timestamp"`
	ChatReference    string `json:"chatReference"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

func CreateMessageTable() {
	_, err := Instance.Exec(
		`CREATE TABLE messages (
			id SERIAL PRIMARY KEY,
			messageReference TEXT NOT NULL,
			text TEXT NOT NULL,
			sender TEXT NOT NULL,
			receiver TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			chatReference REFERENCES chats(chatReference),
			createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	)
	if err != nil {
		log.Fatalf("unable to create table: %v", err)
	}
}

func InsertMessage(message *Message) *Message {
	chat := GetChat(message.ChatReference)
	dbMessage := &Message{}
	if chat == nil {
		chat = InsertChat(Chat{
			ChatReference: message.ChatReference,
			Participants: map[string]bool{
				message.Sender:   true,
				message.Receiver: true,
			},
		})
	}

	err := Instance.QueryRow(
		`INSERT INTO messages (messageReference, text, sender, receiver, timestamp, chatReference)
		VALUES ($1, $2, $3, $4, $5)`,
		message.MessageReference, message.Text, message.Sender, message.Receiver, message.Timestamp, chat.ChatReference,
	).Scan(
		&dbMessage.MessageReference, &dbMessage.Id, &dbMessage.Text, &dbMessage.Sender, &dbMessage.Receiver, &dbMessage.Timestamp,
		&dbMessage.ChatReference, &dbMessage.CreatedAt, &dbMessage.UpdatedAt,
	)
	if err != nil {
		log.Fatal(err)
	}
	return dbMessage
}

func GetMessage(chatReference string, messageReference string) *Message {
	newMessage := &Message{}
	err := Instance.QueryRow(
		`SELECT id, text, sender, receiver, timestamp, chatReference, createdAt, UpdatedAt FROM messages
		WHERE chatReference = $1 AND messageReference = $2`,
		chatReference, messageReference,
	).Scan(
		&newMessage.Id, &newMessage.MessageReference, &newMessage.Text, &newMessage.Sender,
		&newMessage.Receiver, &newMessage.Timestamp, &newMessage.ChatReference, &newMessage.CreatedAt, &newMessage.UpdatedAt,
	)
	if err != nil {
		log.Fatal(err)
	}

	return newMessage
}

func GetAllMessages(chatReference string) []*Message {
	messages := []*Message{}
	rows, err := Instance.Query(
		`SELECT id, text, sender, receiver, timestamp, chatReference, createdAt, UpdatedAt FROM messages WHERE chatReference = $1`,
		chatReference,
	)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		message := &Message{}
		rows.Scan(
			&message.Id, &message.MessageReference, &message.Text, &message.Sender, &message.Receiver, &message.Timestamp,
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
		RETURNING id, messageReference, text, sender, receiver, timestamp, chatReference, createdAt, updatedAt`,
	).Scan(
		&message.Id, &message.MessageReference, &message.Text, &message.Sender, &message.Receiver,
		&message.Timestamp, &message.ChatReference, &message.CreatedAt, &message.UpdatedAt,
	)
	if err != nil {
		log.Fatal(err)
	}
	return message
}

func DeleteAllMessages(chatReference string) []*Message {
	messages := []*Message{}
	rows, err := Instance.Query(
		`DELETE FROM messages WHERE chatReference = $1,
		RETURNING id, messageReference, text, sender, receiver, timestamp, chatReference, createdAt, updatedAt`,
		chatReference,
	)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		message := &Message{}
		rows.Scan(
			&message.Id, &message.MessageReference, &message.Text, &message.Sender, &message.Receiver, &message.Timestamp,
			&message.ChatReference, &message.CreatedAt, &message.UpdatedAt,
		)
		messages = append(messages, message)
	}
	return messages
}

func DropMessagesTable() {
	_, err := Instance.Exec(`DROP TABLE messages`)
	if err != nil {
		log.Fatal(err)
	}
}
