package database

import (
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
	ShouldDelete         bool    `json:"shouldDelete"`
	MessageStatus        *string `json:"messageStatus"`
	PresenceStatus       *string `json:"presenceStatus"`
	CreatedAt            string  `json:"createdAt"`
	UpdatedAt            string  `json:"updatedAt"`
}

type MessageStatusType int

const (
	TYPING MessageStatusType = iota
	NOT_TYPING
)

func (messageStatus MessageStatusType) GetStatus() string {
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
		inserted, insertErr := InsertMessage(*message)
		if insertErr != nil {
			return nil, insertErr
		}
		// For private chats, create a receipt row for the receiver
		if inserted != nil && inserted.ReceiverUsername != nil && *inserted.ReceiverUsername != "" {
			_, _ = InsertReceipt(MessageReceipt{
				MessageReference: inserted.MessageReference,
				Username:         *inserted.ReceiverUsername,
			})
		}
		return inserted, nil
	}

	if message.ShouldDelete {
		_, deleteErr := DeleteMessage(message.MessageReference)
		fmt.Println("message deleted")
		return nil, deleteErr
	}

	// Check receipt status and update if delivery/seen timestamps are provided
	if existingMessage.ReceiverUsername != nil && *existingMessage.ReceiverUsername != "" {
		receipt, _ := GetReceipt(existingMessage.MessageReference, *existingMessage.ReceiverUsername)
		if receipt != nil {
			if receipt.DeliveredTimestamp == nil && message.DeliveredTimestamp != nil {
				UpdateReceiptDelivered(existingMessage.MessageReference, *existingMessage.ReceiverUsername)
			}
			if receipt.SeenTimestamp == nil && message.SeenTimestamp != nil {
				UpdateReceiptSeen(existingMessage.MessageReference, *existingMessage.ReceiverUsername)
			}
		}
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
	case chat.ChatType == ChatTypePrivate && (msg.ReceiverUsername == nil || *msg.ReceiverUsername == ""):
		msgErr = errors.New("message receiver cannot be empty for private chats")
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
		sentTimestamp, chatReference, isReadReceiptEnabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (messageReference) DO NOTHING
		RETURNING id, messageReference, textMessage, senderUsername, receiverUsername,
		sentTimestamp, chatReference, isReadReceiptEnabled, createdAt, updatedAt`,
		msg.MessageReference, msg.TextMessage, msg.SenderUsername, msg.ReceiverUsername,
		msg.SentTimestamp, chat.ChatReference, msg.IsReadReceiptEnabled,
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
			&message.ChatReference, &message.IsReadReceiptEnabled, &message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return &message, nil
	}

	// ON CONFLICT DO NOTHING returns no rows — message already exists
	return nil, nil
}

func GetMessage(chatReference string, messageReference string) (*Message, error) {
	message := Message{}
	rows, err := Instance.Query(
		`SELECT m.id, m.messageReference, m.textMessage, m.senderUsername, m.receiverUsername, m.sentTimestamp,
		m.chatReference, r.deliveredTimestamp, r.seenTimestamp, m.isReadReceiptEnabled, m.createdAt, m.updatedAt
		FROM messages m
		LEFT JOIN message_receipts r ON m.messageReference = r.messageReference
		WHERE m.chatReference = $1 AND m.messageReference = $2`,
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

func GetMessagesWithIncompleteReceipts(chatReference string, username string) ([]Message, error) {
	messages := []Message{}
	rows, err := Instance.Query(
		`SELECT m.id, m.messageReference, m.textMessage, m.senderUsername, m.receiverUsername,
		m.sentTimestamp, m.chatReference, r.deliveredTimestamp, r.seenTimestamp, m.isReadReceiptEnabled, m.createdAt, m.updatedAt
		FROM messages m
		LEFT JOIN message_receipts r ON m.messageReference = r.messageReference AND r.username = $2
		WHERE m.chatReference = $1
		AND m.senderUsername <> $2
		AND (r.deliveredTimestamp IS NULL OR r.seenTimestamp IS NULL OR r.id IS NULL)`,
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
		`SELECT m.id, m.messageReference, m.textMessage, m.senderUsername, m.receiverUsername, m.sentTimestamp,
		m.chatReference, r.deliveredTimestamp, r.seenTimestamp, m.isReadReceiptEnabled, m.createdAt, m.updatedAt
		FROM messages m
		LEFT JOIN message_receipts r ON m.messageReference = r.messageReference
		WHERE m.chatReference = $1`,
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
		chatReference, isReadReceiptEnabled, createdAt, updatedAt`,
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
			&message.IsReadReceiptEnabled,
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
		chatReference, isReadReceiptEnabled, createdAt, updatedAt`,
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
			&message.IsReadReceiptEnabled,
			&message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return []Message{}, scanErr
		}
		messages = append(messages, message)
	}
	return messages, nil
}

// GetChatsWithUnreadForUser returns all chats where the given user is a
// participant and has at least one message with an incomplete receipt
// (i.e. sent by someone else and not yet fully delivered/seen).
func GetChatsWithUnreadForUser(username string) ([]Chat, error) {
	chats := []Chat{}
	rows, err := Instance.Query(
		`SELECT DISTINCT c.id, c.chatReference, c.chatType, c.name, c.createdAt, c.updatedAt
		FROM participants p
		JOIN chats c ON c.chatReference = p.chatReference
		WHERE p.username = $1
		AND EXISTS (
			SELECT 1 FROM messages m
			LEFT JOIN message_receipts r ON m.messageReference = r.messageReference AND r.username = $1
			WHERE m.chatReference = p.chatReference
			AND m.senderUsername <> $1
			AND (r.deliveredTimestamp IS NULL OR r.seenTimestamp IS NULL OR r.id IS NULL)
		)`,
		username,
	)
	if err != nil {
		return chats, err
	}
	defer rows.Close()

	for rows.Next() {
		c := Chat{}
		scanErr := rows.Scan(&c.Id, &c.ChatReference, &c.ChatType, &c.Name, &c.CreatedAt, &c.UpdatedAt)
		if scanErr != nil {
			return nil, scanErr
		}
		chats = append(chats, c)
	}
	return chats, nil
}

// CleanupAllFullyAcknowledgedMessages removes every message across all chats
// that has a fully acknowledged receipt. Group messages have no receipts so
// they are never affected.
func CleanupAllFullyAcknowledgedMessages() error {
	_, err := Instance.Exec(
		`DELETE FROM messages
		 WHERE messageReference IN (
		     SELECT r.messageReference FROM message_receipts r
		     WHERE r.deliveredTimestamp IS NOT NULL
		       AND r.seenTimestamp IS NOT NULL
		 )`,
	)
	return err
}
