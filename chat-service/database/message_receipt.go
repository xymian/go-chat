package database

import (
	"errors"
	"log"
)

type MessageReceipt struct {
	Id                 string  `json:"id"`
	MessageReference   string  `json:"messageReference"`
	Username           string  `json:"username"`
	DeliveredTimestamp *string `json:"deliveredTimestamp"`
	SeenTimestamp      *string `json:"seenTimestamp"`
}

func InsertReceipt(receipt MessageReceipt) (*MessageReceipt, error) {
	newReceipt := &MessageReceipt{}
	if receipt.MessageReference == "" || receipt.Username == "" {
		return nil, errors.New("messageReference and username are required")
	}
	rows, err := Instance.Query(
		`INSERT INTO message_receipts (messageReference, username, deliveredTimestamp, seenTimestamp)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (messageReference, username) DO UPDATE SET
		deliveredTimestamp = COALESCE(message_receipts.deliveredTimestamp, EXCLUDED.deliveredTimestamp),
		seenTimestamp = COALESCE(message_receipts.seenTimestamp, EXCLUDED.seenTimestamp)
		RETURNING id, messageReference, username, deliveredTimestamp, seenTimestamp`,
		receipt.MessageReference, receipt.Username, receipt.DeliveredTimestamp, receipt.SeenTimestamp,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&newReceipt.Id, &newReceipt.MessageReference, &newReceipt.Username,
			&newReceipt.DeliveredTimestamp, &newReceipt.SeenTimestamp,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return newReceipt, nil
	}
	return nil, nil
}

func GetReceipt(messageReference string, username string) (*MessageReceipt, error) {
	receipt := &MessageReceipt{}
	rows, err := Instance.Query(
		`SELECT id, messageReference, username, deliveredTimestamp, seenTimestamp
		FROM message_receipts WHERE messageReference = $1 AND username = $2`,
		messageReference, username,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&receipt.Id, &receipt.MessageReference, &receipt.Username,
			&receipt.DeliveredTimestamp, &receipt.SeenTimestamp,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return receipt, nil
	}
	return nil, nil
}

func UpdateReceiptDelivered(messageReference string, username string) (*MessageReceipt, error) {
	receipt := &MessageReceipt{}
	rows, err := Instance.Query(
		`UPDATE message_receipts SET deliveredTimestamp = NOW()
		WHERE messageReference = $1 AND username = $2
		RETURNING id, messageReference, username, deliveredTimestamp, seenTimestamp`,
		messageReference, username,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&receipt.Id, &receipt.MessageReference, &receipt.Username,
			&receipt.DeliveredTimestamp, &receipt.SeenTimestamp,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return receipt, nil
	}
	return nil, nil
}

func UpdateReceiptSeen(messageReference string, username string) (*MessageReceipt, error) {
	receipt := &MessageReceipt{}
	rows, err := Instance.Query(
		`UPDATE message_receipts SET seenTimestamp = NOW()
		WHERE messageReference = $1 AND username = $2
		RETURNING id, messageReference, username, deliveredTimestamp, seenTimestamp`,
		messageReference, username,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&receipt.Id, &receipt.MessageReference, &receipt.Username,
			&receipt.DeliveredTimestamp, &receipt.SeenTimestamp,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return receipt, nil
	}
	return nil, nil
}
