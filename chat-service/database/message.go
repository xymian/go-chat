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
	ReceiverUsername     *string `json:"receiverUsername"`
	SentTimestamp        string  `json:"sentTimestamp"`
	ChatReference        string  `json:"chatReference"`
	DeliveredTimestamp   *string `json:"deliveredTimestamp"`
	SeenTimestamp        *string `json:"seenTimestamp"`
	IsReadReceiptEnabled bool    `json:"isReadReceiptEnabled"`
	IsBackedUp           bool    `json:"isBackedUp"`
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
		if !existingMessage.IsBackedUp {
			_, deleteErr := DeleteMessage(message.MessageReference)
			fmt.Println("message deleted")
			return nil, deleteErr
		}
		return nil, nil
	}

	if existingMessage.DeliveredTimestamp == nil {
		if existingMessage.IsBackedUp == true {
			message.IsBackedUp = true
		}
		return InsertMessage(*message)
	}
	if existingMessage.SeenTimestamp == nil {
		message.DeliveredTimestamp = existingMessage.DeliveredTimestamp
		if existingMessage.IsBackedUp == true {
			message.IsBackedUp = true
		}
		return InsertMessage(*message)
	}

	if message.IsBackedUp == true {
		existingMessage.IsBackedUp = true
		return InsertMessage(*existingMessage)
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
	case !chat.IsGroup && (msg.ReceiverUsername == nil || *msg.ReceiverUsername == ""):
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
		sentTimestamp, chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, is_backed_up)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (messageReference) DO UPDATE SET
		deliveredTimestamp = EXCLUDED.deliveredTimestamp,
		seenTimestamp = EXCLUDED.seenTimestamp,
		is_backed_up = EXCLUDED.is_backed_up
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,
		sentTimestamp, chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, is_backed_up, createdAt, updatedAt`,
		msg.MessageReference, msg.TextMessage, msg.SenderUsername, msg.ReceiverUsername,
		msg.SentTimestamp, chat.ChatReference, msg.DeliveredTimestamp, msg.SeenTimestamp, msg.IsReadReceiptEnabled, msg.IsBackedUp,
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
			&message.IsReadReceiptEnabled, &message.IsBackedUp, &message.CreatedAt, &message.UpdatedAt,
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
		chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, is_backed_up, createdAt, UpdatedAt FROM messages
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
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled, &message.IsBackedUp, &message.CreatedAt, &message.UpdatedAt,
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

	isGroup, err := IsGroupChat(chatReference)
	if err != nil {
		return nil, err
	}
	if isGroup {
		return AcknowledgeGroupMessages(chatReference, username, from, to)
	}

	messages := []Message{}

	ctx := context.Background()
	txn, txnErr := Instance.BeginTx(ctx, nil)
	if txnErr != nil {
		return nil, txnErr
	}
	defer txn.Rollback()

	rows, queryErr := txn.Query(
		`UPDATE messages SET seenTimestamp = NOW()
		WHERE chatReference = $1
		AND senderUsername <> $2
		AND sentTimestamp BETWEEN $3 AND $4
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,
		sentTimestamp, chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, is_backed_up, createdAt, updatedAt`,
		chatReference, username, from, to,
	)
	if queryErr != nil {
		return []Message{}, queryErr
	}
	defer rows.Close()

	for rows.Next() {
		message := Message{}
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.SentTimestamp, &message.ChatReference,
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled, &message.IsBackedUp, &message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return []Message{}, scanErr
		}
		messages = append(messages, message)
	}

	if e := txn.Commit(); e != nil {
		return []Message{}, e
	}

	// cleanup fully‐acknowledged, unbacked messages
	_ = cleanupFullyAcknowledgedMessages(chatReference)

	return messages, nil
}

func GetAllUnacknowledgedMessages(chatReference string, username string) ([]Message, error) {
	isGroup, err := IsGroupChat(chatReference)
	if err != nil {
		return []Message{}, nil
	}
	if isGroup {
		return getUnacknowledgedGroupMessages(chatReference, username)
	}

	messages := []Message{}
	rows, queryErr := Instance.Query(
		`SELECT id, messageReference, textMessage, senderUsername, receiverUsername,
		sentTimestamp, chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, is_backed_up, createdAt, updatedAt
		FROM messages WHERE chatReference = $1 AND (deliveredTimestamp IS NULL OR seenTimestamp IS NULL) AND senderUsername <> $2`,
		chatReference, username,
	)
	if queryErr != nil {
		return []Message{}, nil
	}
	defer rows.Close()

	for rows.Next() {
		message := Message{}
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage, &message.SenderUsername,
			&message.ReceiverUsername, &message.SentTimestamp, &message.ChatReference,
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled, &message.IsBackedUp, &message.CreatedAt, &message.UpdatedAt,
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
		chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, is_backed_up, createdAt, updatedAt`,
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
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled, &message.IsBackedUp,
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
		chatReference, deliveredTimestamp, seenTimestamp, isReadReceiptEnabled, is_backed_up, createdAt, updatedAt`,
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
			&message.DeliveredTimestamp, &message.SeenTimestamp, &message.IsReadReceiptEnabled, &message.IsBackedUp,
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
// both delivery and seen timestamps and are not marked for backup.
func cleanupFullyAcknowledgedMessages(chatReference string) error {
	_, err := Instance.Exec(
		`DELETE FROM messages
		 WHERE chatReference = $1
		   AND deliveredTimestamp IS NOT NULL
		   AND seenTimestamp IS NOT NULL
		   AND is_backed_up = FALSE`,
		chatReference,
	)
	return err
}

// CleanupAllFullyAcknowledgedMessages removes every DM message in every chat
// that has been fully acknowledged and is not backed up, and removes group
// messages where all participants have seen them.
func CleanupAllFullyAcknowledgedMessages() error {
	_, err := Instance.Exec(
		`DELETE FROM messages
		 WHERE receiverUsername IS NOT NULL
		   AND deliveredTimestamp IS NOT NULL
		   AND seenTimestamp IS NOT NULL
		   AND is_backed_up = FALSE`,
	)
	if err != nil {
		return err
	}

	// Group messages: delete where all non-sender participants have a seenAt receipt.
	_, err = Instance.Exec(
		`DELETE FROM messages
		 WHERE receiverUsername IS NULL
		   AND is_backed_up = FALSE
		   AND NOT EXISTS (
		       SELECT 1
		       FROM participants p
		       LEFT JOIN message_receipts mr
		           ON messages.messageReference = mr.messageReference
		           AND mr.username = p.username
		       WHERE p.chatReference = messages.chatReference
		         AND p.username <> messages.senderUsername
		         AND mr.seenAt IS NULL
		   )`,
	)
	return err
}
