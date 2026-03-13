package database

import "context"

// MessageReceipt tracks per-user delivery and seen status for group messages.
type MessageReceipt struct {
	Id               string  `json:"id"`
	MessageReference string  `json:"messageReference"`
	Username         string  `json:"username"`
	DeliveredAt      *string `json:"deliveredAt"`
	SeenAt           *string `json:"seenAt"`
	CreatedAt        string  `json:"createdAt"`
}

// MarkGroupMessageDelivered records that username received the given message.
// Uses INSERT ... ON CONFLICT DO UPDATE so it's safe to call multiple times.
func MarkGroupMessageDelivered(messageReference string, username string) error {
	_, err := Instance.Exec(
		`INSERT INTO message_receipts (messageReference, username, deliveredAt)
		VALUES ($1, $2, NOW())
		ON CONFLICT (messageReference, username)
		DO UPDATE SET deliveredAt = COALESCE(message_receipts.deliveredAt, NOW())`,
		messageReference, username,
	)
	return err
}

// AcknowledgeGroupMessages marks as seen all group messages in chatReference
// sent by someone other than username whose sentTimestamp is between from and to.
// Returns the updated messages.
func AcknowledgeGroupMessages(
	chatReference string, username string, from string, to string,
) ([]Message, error) {
	messages := []Message{}

	ctx := context.Background()
	txn, err := Instance.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	// Upsert a receipt for every qualifying message.
	_, err = txn.Exec(
		`INSERT INTO message_receipts (messageReference, username, deliveredAt, seenAt)
		SELECT m.messageReference, $2, NOW(), NOW()
		FROM messages m
		WHERE m.chatReference = $1
		  AND m.senderUsername <> $2
		  AND m.sentTimestamp BETWEEN $3 AND $4
		ON CONFLICT (messageReference, username)
		DO UPDATE SET
		    deliveredAt = COALESCE(message_receipts.deliveredAt, NOW()),
		    seenAt      = COALESCE(message_receipts.seenAt, NOW())`,
		chatReference, username, from, to,
	)
	if err != nil {
		return nil, err
	}

	msgRows, err := txn.Query(
		`SELECT m.id, m.messageReference, m.textMessage, m.senderUsername,
		m.receiverUsername, m.sentTimestamp, m.chatReference,
		m.deliveredTimestamp, m.seenTimestamp, m.isReadReceiptEnabled, m.is_backed_up,
		m.createdAt, m.updatedAt
		FROM messages m
		WHERE m.chatReference = $1
		  AND m.senderUsername <> $2
		  AND m.sentTimestamp BETWEEN $3 AND $4`,
		chatReference, username, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer msgRows.Close()

	for msgRows.Next() {
		message := Message{}
		scanErr := msgRows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage,
			&message.SenderUsername, &message.ReceiverUsername, &message.SentTimestamp,
			&message.ChatReference, &message.DeliveredTimestamp, &message.SeenTimestamp,
			&message.IsReadReceiptEnabled, &message.IsBackedUp,
			&message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		messages = append(messages, message)
	}

	if err := txn.Commit(); err != nil {
		return nil, err
	}

	_ = cleanupFullyAcknowledgedGroupMessages(chatReference)
	return messages, nil
}

// getUnacknowledgedGroupMessages returns group messages in chatReference that
// username has not yet seen (no receipt row, or receipt with NULL seenAt).
func getUnacknowledgedGroupMessages(chatReference string, username string) ([]Message, error) {
	messages := []Message{}
	rows, err := Instance.Query(
		`SELECT m.id, m.messageReference, m.textMessage, m.senderUsername,
		m.receiverUsername, m.sentTimestamp, m.chatReference,
		m.deliveredTimestamp, m.seenTimestamp, m.isReadReceiptEnabled, m.is_backed_up,
		m.createdAt, m.updatedAt
		FROM messages m
		LEFT JOIN message_receipts mr
		    ON m.messageReference = mr.messageReference AND mr.username = $2
		WHERE m.chatReference = $1
		  AND m.senderUsername <> $2
		  AND m.receiverUsername IS NULL
		  AND mr.seenAt IS NULL`,
		chatReference, username,
	)
	if err != nil {
		return messages, err
	}
	defer rows.Close()

	for rows.Next() {
		message := Message{}
		scanErr := rows.Scan(
			&message.Id, &message.MessageReference, &message.TextMessage,
			&message.SenderUsername, &message.ReceiverUsername, &message.SentTimestamp,
			&message.ChatReference, &message.DeliveredTimestamp, &message.SeenTimestamp,
			&message.IsReadReceiptEnabled, &message.IsBackedUp,
			&message.CreatedAt, &message.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		messages = append(messages, message)
	}
	return messages, nil
}

// cleanupFullyAcknowledgedGroupMessages deletes group messages in chatReference
// where every non-sender participant has a seenAt receipt and the message is not
// backed up.
func cleanupFullyAcknowledgedGroupMessages(chatReference string) error {
	_, err := Instance.Exec(
		`DELETE FROM messages
		WHERE chatReference = $1
		  AND receiverUsername IS NULL
		  AND is_backed_up = FALSE
		  AND NOT EXISTS (
		      SELECT 1
		      FROM participants p
		      LEFT JOIN message_receipts mr
		          ON messages.messageReference = mr.messageReference
		          AND mr.username = p.username
		      WHERE p.chatReference = $1
		        AND p.username <> messages.senderUsername
		        AND mr.seenAt IS NULL
		  )`,
		chatReference,
	)
	return err
}
