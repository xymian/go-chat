package database

import (
	"context"
	"errors"
	"log"

	"github.com/te6lim/go-chat/chat-service/models"
)

type Message struct {
	Id                   string  `json:"id"`
	MessageReference     string  `json:"messageReference"`
	TextMessage          string  `json:"textMessage"`
	SenderUsername       string  `json:"senderUsername"`
	ReceiverUsername     string  `json:"receiverUsername"`
	MessageTimestamp     string  `json:"messageTimestamp"`
	ChatReference        string  `json:"chatReference"`
	Ack                  bool    `json:"ack"`
	DeliveredTimestamp   *string `json:"deliveredTimestamp"`
	SeenTimestamp        *string `json:"seenTimestamp"`
	IsReadReceiptEnabled bool    `json:"isReadReceiptEnabled"`
	MessageStatus        *string `json:"messageStatus"`
	PresenceStatus       *string `json:"presenceStatus"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
}

type MessageStatus int

const (
	TYPING MessageStatus = iota
	NOT_TYPING
)

func (messageStatus MessageStatus) GetStatus() string {
	switch messageStatus {
	case TYPING:
		return "TYPING"
	case NOT_TYPING:
		return "NOT_TYPING"
	}
	return ""
}

func MarkMessagesAsDelivered(messagesDetails models.DeliverMessages) ([]Message, error) {
	messages := []Message{}
	ctx := context.Background()
	txn, err := Instance.BeginTx(ctx, nil)
	if err != nil {
		return messages, err
	}
	defer txn.Rollback()

	for _, detail := range messagesDetails.MessagesDetails {
		message := Message{}
		rows, err := txn.Query(
			`UPDATE messages SET deliveredTimestamp = $1 WHERE senderUsername = $2 AND chatReference = $3 AND
			messageReference = $4 AND messageTimestamp = $5 AND seenTimestamp = $6 AND isReadReceiptEnabled = $7
			RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,messageTimestamp,
			chatReference, ack, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt`,
			detail.DeliveredTimestamp, messagesDetails.Sender, messagesDetails.ChatReference,
			detail.MessageReference, detail.SentTimestamp, detail.ReadTimestamp, detail.IsReadReceiptEnabled,
		)

		if err != nil {
			return []Message{}, nil
		}

		for rows.Next() {
			scanErr := rows.Scan(
				&message.Id, &message.MessageReference, &message.TextMessage,
				&message.SenderUsername, &message.ReceiverUsername, &message.MessageTimestamp,
				&message.ChatReference, &message.Ack, &message.DeliveredTimestamp, &message.SeenTimestamp,
				&message.IsReadReceiptEnabled, &message.CreatedAt, &message.UpdatedAt,
			)

			if scanErr != nil {
				rows.Close()
				return nil, scanErr
			}
		}

		messages = append(messages, message)
	}
	e := txn.Commit()
	if e != nil {
		return []Message{}, e
	}
	return messages, nil
}

func MaybeInsertAndReturnMostUpToDateMessage(message *Message) (*Message, error) {
	existingMessage, err := GetMessage(message.ChatReference, message.MessageReference)

	if existingMessage == nil {
		return InsertMessage(*message)
	}
	if existingMessage.DeliveredTimestamp == nil {
		return InsertMessage(*message)
	}
	if existingMessage.SeenTimestamp == nil {
		message.Ack = existingMessage.Ack
		message.DeliveredTimestamp = existingMessage.DeliveredTimestamp
		return InsertMessage(*message)
	}

	return existingMessage, err
}

func InsertMessage(msg Message) (*Message, error) {
	chat, err := GetChat(msg.ChatReference)
	if err != nil {
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
		println(msgErr.Error())
		return nil, msgErr
	}
	message := Message{}
	rows, queryErr := Instance.Query(
		`INSERT INTO messages (messageReference, textMessage, senderUsername, receiverUsername,
		messageTimestamp, chatReference, ack, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (messageReference) DO UPDATE SET
		ack = EXCLUDED.ack,
		deliveredTimestamp = EXCLUDED.deliveredTimestamp,
		seenTimestamp = EXCLUDED.seenTimestamp
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,
		messageTimestamp, chatReference, ack, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt`,
		msg.MessageReference, msg.TextMessage, msg.SenderUsername, msg.ReceiverUsername,
		msg.MessageTimestamp, chat.ChatReference, msg.Ack, msg.DeliveredTimestamp, msg.SeenTimestamp, msg.IsReadReceiptEnabled,
	)

	if queryErr != nil {
		log.Fatal(queryErr)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage,
			&message.SenderUsername, &message.ReceiverUsername, &message.MessageTimestamp,
			&message.ChatReference, &message.Ack, &message.DeliveredTimestamp, &message.SeenTimestamp,
			&message.IsReadReceiptEnabled, &message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return &message, nil
	}

	return nil, nil
}

func GetMessage(chatReference string, messageReference string) (*Message, error) {
	message := Message{}
	rows, err := Instance.Query(
		`SELECT id, messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp,
		chatReference, ack, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, UpdatedAt FROM messages
		WHERE chatReference = $1 AND messageReference = $2`,
		chatReference, messageReference,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference, &message.Ack,
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled, &message.CreatedAt, &message.UpdatedAt,
		)

		if scanErr != nil {
			return nil, scanErr
		}

		return &message, nil
	}
	return nil, nil
}

func AcknowledgeMessages(
	chatReference string, username string, from string, to string) ([]Message, error) {
	messages := []Message{}

	rows, err := Instance.Query(
		`UPDATE messages SET ack = $1 WHERE senderUsername <> $2 AND chatReference = $3
		AND ack = $4 AND messageTimestamp BETWEEN $5 AND $6
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,
		messageTimestamp, chatReference, ack, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt`,
		"true", username, chatReference, "false", from, to,
	)
	if err != nil {
		return []Message{}, nil
	}

	defer rows.Close()

	for rows.Next() {
		message := Message{}
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference, &message.Ack,
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled, &message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return []Message{}, scanErr
		}

		messages = append(messages, message)
	}

	return messages, nil
}

func GetAllUnacknowledgedMessages(chatReference string, username string) ([]Message, error) {
	messages := []Message{}
	rows, err := Instance.Query(
		`SELECT id, messageReference, textMessage, senderUsername, receiverUsername,
		messageTimestamp, chatReference, ack, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt
		FROM messages WHERE chatReference = $1 AND deliveredTimestamp IS NULL AND senderUsername <> $2`,
		chatReference, username,
	)
	if err != nil {
		return []Message{}, nil
	}

	defer rows.Close()

	for rows.Next() {
		message := Message{}
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference, &message.Ack,
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled, &message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func GetAllMessages(chatReference string) ([]Message, error) {
	messages := []Message{}
	rows, err := Instance.Query(
		`SELECT id, messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp,
		chatReference, ack, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt FROM messages WHERE chatReference = $1`,
		chatReference,
	)
	if err != nil {
		return []Message{}, nil
	}

	defer rows.Close()

	for rows.Next() {
		message := Message{}
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference, &message.Ack,
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled, &message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func DeleteMessage(messageReference string) (*Message, error) {
	message := Message{}
	rows, err := Instance.Query(
		`DELETE FROM messages WHERE messageReference = $1
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp,
		chatReference, ack, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt`,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference, &message.Ack,
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled,
			&message.CreatedAt, &message.UpdatedAt,
		)

		if scanErr != nil {
			return nil, scanErr
		}

		return &message, nil
	}

	return nil, nil
}

func DeleteAllMessages(chatReference string) ([]Message, error) {
	messages := []Message{}
	rows, err := Instance.Query(
		`DELETE FROM messages WHERE chatReference = $1,
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername, messageTimestamp,
		chatReference, ack, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt`,
		chatReference,
	)
	if err != nil {
		return []Message{}, nil
	}

	defer rows.Close()

	for rows.Next() {
		message := Message{}
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.MessageTimestamp, &message.ChatReference,
			&message.Ack, &message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled,
			&message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return []Message{}, scanErr
		}
		messages = append(messages, message)
	}
	return messages, nil
}
