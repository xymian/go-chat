package database

import (
	"context"
	"errors"
	"log"

	"github.com/te6lim/go-chat/models"
)

type Message struct {
	Id               string `json:"id"`
	MessageReference string `json:"messageReference"`
	TextMessage      string `json:"textMessage"`
	SenderUsername   string `json:"senderUsername"`
	ReceiverUsername string `json:"receiverUsername"`
	MessageTimestamp string `json:"messageTimestamp"`
	ChatReference    string `json:"chatReference"`
	Ack              bool   `json:"ack"`
	Delivered        bool   `json:"delivered"`
	Seen             bool   `json:"seen"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

func MarkMessagesAsDelivered(details models.DeliverMessage) (*[]Message, error) {
	messages := []Message{}
	ctx := context.Background()
	txn, err := Instance.BeginTx(ctx, nil)
	if err != nil {
		return &messages, err
	}
	defer txn.Rollback()

	for messageRef, timestamp := range details.MessageDetails {
		message := Message{}
		txn.QueryRow(
			`UPDATE messages SET delivered = $1 WHERE senderUsername = $2 AND chatReference = $3 AND
			messageReference = $4 AND messageTimestamp = $5 AND seen = $6
			RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,messageTimestamp,
			chatReference, ack, delivered, seen, createdAt, updatedAt`,
			true, details.Sender, details.ChatReference, messageRef, timestamp, details.Seen,
		).Scan(
			&message.Id, &message.MessageReference, &message.TextMessage,
			&message.SenderUsername, &message.ReceiverUsername, &message.MessageTimestamp,
			&message.ChatReference, &message.Ack, &message.Delivered, &message.Seen, &message.CreatedAt, &message.UpdatedAt,
		)

		messages = append(messages, message)
	}
	e := txn.Commit()
	if e != nil {
		return &[]Message{}, e
	}
	return &messages, nil
}

func InsertMessage(msg Message) (*Message, error) {
	chat := GetChat(msg.ChatReference)
	if chat == nil {
		return nil, errors.New("chat reference for this message does not exits")
	}
	var msgErr error = nil
	switch {
	case msg.MessageReference == "":
		msgErr = errors.New("message reference cannot be empty")
	case msg.TextMessage == "":
		msgErr = errors.New("message text cannot be empty")
	case msg.SenderUsername == "":
		msgErr = errors.New("message sender cannot be empty")
	case msg.ReceiverUsername == "":
		msgErr = errors.New("message receiver cannot be empty")
	case msg.MessageTimestamp == "":
		msgErr = errors.New("message timestamp cannot be empty")
	}

	if msgErr != nil {
		println("error from one of the cases")
		return nil, msgErr
	}
	message := Message{}
	Instance.QueryRow(
		`INSERT INTO messages (messageReference, textMessage, senderUsername, receiverUsername,
		messageTimestamp, chatReference, ack, delivered, seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,
		messageTimestamp, chatReference, ack, delivered, seen, createdAt, updatedAt`,
		msg.MessageReference, msg.TextMessage, msg.SenderUsername, msg.ReceiverUsername,
		msg.MessageTimestamp, chat.ChatReference, msg.Ack, msg.Delivered, msg.Seen,
	).Scan(
		&message.Id, &message.MessageReference, &message.TextMessage,
		&message.SenderUsername, &message.ReceiverUsername, &message.MessageTimestamp,
		&message.ChatReference, message.Ack, &message.Delivered, &message.Seen, &message.CreatedAt, &message.UpdatedAt,
	)
	return &message, msgErr
}

func GetMessage(chatReference string, messageReference string) *Message {
	message := Message{}
	Instance.QueryRow(
		`SELECT id, textMessage, senderUsername, receiverUsername, messageTimestamp,
		chatReference, ack, delivered, seen, createdAt, UpdatedAt FROM messages
		WHERE chatReference = $1 AND messageReference = $2`,
		chatReference, messageReference,
	).Scan(
		&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
		&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference, &message.Ack,
		&message.Delivered, &message.Seen, &message.CreatedAt, &message.UpdatedAt,
	)

	return &message
}

func AcknowledgeMessages(
	chatReference string, username string, from string, to string) []*Message {
	messages := []*Message{}

	rows, err := Instance.Query(
		`UPDATE messages SET ack = $1 WHERE senderUsername <> $2 AND chatReference = $3
		AND ack = $4 AND messageTimestamp BETWEEN $5 AND $6
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,
		messageTimestamp, chatReference, ack, delivered, seen, createdAt, updatedAt`,
		"true", username, chatReference, "false", from, to,
	)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	for rows.Next() {
		message := Message{}
		rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference, &message.Ack,
			&message.Delivered, &message.Seen, &message.CreatedAt, &message.UpdatedAt,
		)
		messages = append(messages, &message)
	}
	println("message size: ", len(messages))
	return messages
}

func GetAllUnacknowledgedMessages(chatReference string, username string) []*Message {
	messages := []*Message{}
	rows, err := Instance.Query(
		`SELECT id, messageReference, textMessage, senderUsername, receiverUsername,
		messageTimestamp, chatReference, ack, delivered, seen, createdAt, updatedAt
		FROM messages WHERE chatReference = $1 AND ack = $2 AND senderUsername <> $3`,
		chatReference, "false", username,
	)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	for rows.Next() {
		message := Message{}
		rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference, &message.Ack,
			&message.Delivered, &message.Seen, &message.CreatedAt, &message.UpdatedAt,
		)
		messages = append(messages, &message)
	}
	return messages
}

func GetAllMessages(chatReference string) []*Message {
	messages := []*Message{}
	rows, err := Instance.Query(
		`SELECT id, messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp,
		chatReference, ack, delivered, seen, createdAt, updatedAt FROM messages WHERE chatReference = $1`,
		chatReference,
	)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	for rows.Next() {
		message := Message{}
		rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference, &message.Ack,
			&message.Delivered, &message.Seen, &message.CreatedAt, &message.UpdatedAt,
		)
		messages = append(messages, &message)
	}
	return messages
}

func DeleteMessage(messageReference string) *Message {
	message := Message{}
	Instance.QueryRow(
		`DELETE FROM messages WHERE messageReference = $1
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp,
		chatReference, ack, delivered, seen, createdAt, updatedAt`,
	).Scan(
		&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername, &message.ReceiverUsername,
		&message.MessageTimestamp, &message.ChatReference, &message.Ack, &message.Delivered,
		&message.Seen, &message.CreatedAt, &message.UpdatedAt,
	)

	return &message
}

func DeleteAllMessages(chatReference string) []*Message {
	messages := []*Message{}
	rows, err := Instance.Query(
		`DELETE FROM messages WHERE chatReference = $1,
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp,
		chatReference, ack, delivered, seen, createdAt, updatedAt`,
		chatReference,
	)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		message := Message{}
		rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference,
			&message.Ack, &message.Delivered, &message.Seen, &message.CreatedAt, &message.UpdatedAt,
		)
		messages = append(messages, &message)
	}
	return messages
}
