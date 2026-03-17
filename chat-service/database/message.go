package database

import (
	"context"
	"errors"
	"fmt"
	"log"
)

type Message struct {
	Id                   string  `json:"id"`
	MessageReference     string  `json:"messageReference"`
	TextMessage          string  `json:"textMessage"`
	SenderUsername       string  `json:"senderUsername"`
	ReceiverUsername     string  `json:"receiverUsername"`
	SentTimestamp        string  `json:"sentTimestamp"`
	ChatReference        string  `json:"chatReference"`
	DeliveredTimestamp   *string `json:"deliveredTimestamp"`
	SeenTimestamp        *string `json:"seenTimestamp"`
	IsReadReceiptEnabled bool    `json:"isReadReceiptEnabled"`
	ShouldDelete         bool    `json:"shouldDelete"`
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

func MaybeInsertAndReturnMostUpToDateMessage(message *Message) (*Message, error) {
	existingMessage, err := GetMessage(message.ChatReference, message.MessageReference)

	if existingMessage == nil {
		return InsertMessage(*message)
	}

	if message.ShouldDelete {
		_, deleteErr := DeleteMessage(message.MessageReference)
		fmt.Println("message deleted")
		return nil, deleteErr
	}

	if existingMessage.DeliveredTimestamp == nil {
		return InsertMessage(*message)
	}
	if existingMessage.SeenTimestamp == nil {
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
	case msg.SentTimestamp == "":
		msgErr = errors.New("message timestamp cannot be empty")
	}

	if msgErr != nil {
		println(msgErr.Error())
		return nil, msgErr
	}
	message := Message{}
	rows, queryErr := Instance.Query(
		`INSERT INTO messages (messageReference, textMessage, senderUsername, receiverUsername,
		sentTimestamp, chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (messageReference) DO UPDATE SET
		deliveredTimestamp = EXCLUDED.deliveredTimestamp,
		seenTimestamp = EXCLUDED.seenTimestamp
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,
		sentTimestamp, chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt`,
		msg.MessageReference, msg.TextMessage, msg.SenderUsername, msg.ReceiverUsername,
		msg.SentTimestamp, chat.ChatReference, msg.DeliveredTimestamp, msg.SeenTimestamp, msg.IsReadReceiptEnabled,
	)

	if queryErr != nil {
		log.Fatal(queryErr)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage,
			&message.SenderUsername, &message.ReceiverUsername, &message.SentTimestamp,
			&message.ChatReference, &message.DeliveredTimestamp, &message.SeenTimestamp,
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
		`SELECT id, messageReference, textMessage, senderUsername, receiverUsername, sentTimestamp,
		chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, UpdatedAt FROM messages
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
			&message.ReceiverUsername, &message.SentTimestamp, &message.ChatReference,
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

	ctx := context.Background()
	txn, err := Instance.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	rows, err := txn.Query(
		`UPDATE messages SET seenTimestamp = NOW()
		WHERE chatReference = $1
		AND senderUsername <> $2
		AND sentTimestamp BETWEEN $3 AND $4
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,
		sentTimestamp, chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt`,
		chatReference, username, from, to,
	)
	if err != nil {
		return []Message{}, err
	}
	defer rows.Close()

	for rows.Next() {
		message := Message{}
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.SentTimestamp, &message.ChatReference,
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled, &message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return []Message{}, scanErr
		}
		messages = append(messages, message)
	}

	e := txn.Commit()
	if e != nil {
		return []Message{}, e
	}

	// cleanup fully acknowledged messages
	_ = cleanupFullyAcknowledgedMessages(chatReference)

	return messages, nil
}

func GetAllUnacknowledgedMessages(chatReference string, username string) ([]Message, error) {
	messages := []Message{}
	rows, err := Instance.Query(
		`SELECT id, messageReference, textMessage, senderUsername, receiverUsername,
		sentTimestamp, chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt
		FROM messages WHERE chatReference = $1 AND (deliveredTimestamp IS NULL OR seenTimestamp IS NULL) AND senderUsername <> $2`,
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
			&message.ReceiverUsername, &message.SentTimestamp, &message.ChatReference,
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
		`SELECT id, messageReference, textMessage, senderUsername, receiverUsername, sentTimestamp,
		chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt FROM messages WHERE chatReference = $1`,

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
			&message.ReceiverUsername, &message.SentTimestamp, &message.ChatReference,
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
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername, sentTimestamp,
		chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt`,
		messageReference,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.SentTimestamp, &message.ChatReference,
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
		`DELETE FROM messages WHERE chatReference = $1
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername, sentTimestamp,
		chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, createdAt, updatedAt`,
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
			&message.ReceiverUsername, &message.SentTimestamp, &message.ChatReference,
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled,
			&message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return []Message{}, scanErr
		}
		messages = append(messages, message)
	}
	return messages, nil
}

// cleanupFullyAcknowledgedMessages deletes all messages inside a chat that have
// both delivery and seen timestamps.
func cleanupFullyAcknowledgedMessages(chatReference string) error {
	_, err := Instance.Exec(
		`DELETE FROM messages
		 WHERE chatReference = $1
		   AND deliveredTimestamp IS NOT NULL
		   AND seenTimestamp IS NOT NULL`,
		chatReference,
	)
	return err
}

// CleanupAllFullyAcknowledgedMessages removes every message in every chat that
// has been fully acknowledged. This can be run periodically as a safety net.
func CleanupAllFullyAcknowledgedMessages() error {
	_, err := Instance.Exec(
		`DELETE FROM messages
		 WHERE deliveredTimestamp IS NOT NULL
		   AND seenTimestamp IS NOT NULL`,
	)
	return err
}
