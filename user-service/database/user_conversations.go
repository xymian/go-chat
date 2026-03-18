package database

import (
	"log"
)

type UserConversation struct {
	Id            int64  `json:"id"`
	UserId        int64  `json:"userId"`
	ChatReference string `json:"chatReference"`
	ChatType      string `json:"chatType"`
	OtherUserId   int64  `json:"otherUserId"`
	Visible       bool   `json:"visible"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

func AddUserConversation(userId int64, chatReference string, chatType string, otherUserId int64) (*UserConversation, error) {
	conv := &UserConversation{}
	rows, err := Instance.Query(
		`INSERT INTO user_conversations(userId, chatReference, chatType, otherUserId, visible)
		VALUES($1, $2, $3, $4, TRUE)
		ON CONFLICT(userId, chatReference) DO UPDATE SET visible = TRUE, updatedAt = CURRENT_TIMESTAMP
		RETURNING id, userId, chatReference, chatType, otherUserId, visible, createdAt, updatedAt`,
		userId, chatReference, chatType, otherUserId,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&conv.Id, &conv.UserId, &conv.ChatReference, &conv.ChatType,
			&conv.OtherUserId, &conv.Visible, &conv.CreatedAt, &conv.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return conv, nil
	}
	return nil, nil
}

func RemoveUserConversation(userId int64, chatReference string) error {
	_, err := Instance.Exec(
		`UPDATE user_conversations SET visible = FALSE, updatedAt = CURRENT_TIMESTAMP
		WHERE userId = $1 AND chatReference = $2`,
		userId, chatReference,
	)
	return err
}

func GetUserConversations(userId int64) ([]UserConversation, error) {
	convs := []UserConversation{}
	rows, err := Instance.Query(
		`SELECT id, userId, chatReference, chatType, otherUserId, visible, createdAt, updatedAt
		FROM user_conversations
		WHERE userId = $1 AND visible = TRUE`,
		userId,
	)
	if err != nil {
		return convs, err
	}

	defer rows.Close()

	for rows.Next() {
		conv := UserConversation{}
		scanErr := rows.Scan(
			&conv.Id, &conv.UserId, &conv.ChatReference, &conv.ChatType,
			&conv.OtherUserId, &conv.Visible, &conv.CreatedAt, &conv.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		convs = append(convs, conv)
	}
	return convs, nil
}
