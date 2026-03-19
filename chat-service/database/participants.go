package database

import (
	"errors"
	"log"
)

type ParticipantStatus string

const (
	ParticipantStatusPending  ParticipantStatus = "PENDING"
	ParticipantStatusAccepted ParticipantStatus = "ACCEPTED"
)

type Participant struct {
	Id            string            `json:"id"`
	Username      string            `json:"username"`
	ChatReference string            `json:"chatReference"`
	Status        ParticipantStatus `json:"status"`
	CreatedAt     string            `json:"createdAt"`
	UpdatedAt     string            `json:"updatedAt"`
}

func InsertParticipant(participant Participant) (*Participant, error) {
	newParticipant := &Participant{}
	if len(participant.Username) == 0 || len(participant.ChatReference) == 0 {
		return nil, errors.New("invalid participant")
	}
	status := participant.Status
	if status == "" {
		status = ParticipantStatusAccepted
	}
	rows, err := Instance.Query(
		`INSERT INTO participants (username, chatReference, status) VALUES ($1, $2, $3) RETURNING id, username, chatReference, status, createdAt, updatedAt`,
		participant.Username, participant.ChatReference, status,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(
			&newParticipant.Id, &newParticipant.Username, &newParticipant.ChatReference,
			&newParticipant.Status, &newParticipant.CreatedAt, &newParticipant.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return newParticipant, nil
	}
	return nil, nil
}

func GetParticipantsInChat(chatReference string) ([]Participant, error) {
	participants := []Participant{}
	rows, err := Instance.Query(
		`SELECT id, username, chatReference, status, createdAt, updatedAt FROM participants WHERE chatReference = $1`,
		chatReference,
	)
	if err != nil {
		return []Participant{}, nil
	}

	defer rows.Close()

	for rows.Next() {
		participant := &Participant{}
		scanErr := rows.Scan(
			&participant.Id, &participant.Username, &participant.ChatReference,
			&participant.Status, &participant.CreatedAt, &participant.UpdatedAt,
		)
		if scanErr != nil {
			return []Participant{}, scanErr
		}
		participants = append(participants, *participant)
	}
	return participants, nil
}

func GetParticipant(username string, chatReference string) (*Participant, error) {
	participant := &Participant{}
	rows, err := Instance.Query(
		`SELECT id, username, chatReference, status, createdAt, updatedAt FROM participants WHERE username = $1 AND chatReference = $2`,
		username, chatReference,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(
			&participant.Id, &participant.Username, &participant.ChatReference,
			&participant.Status, &participant.CreatedAt, &participant.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return participant, nil
	}
	return nil, nil
}

func UpdateParticipantStatus(username string, chatReference string, status ParticipantStatus) (*Participant, error) {
	participant := &Participant{}
	rows, err := Instance.Query(
		`UPDATE participants SET status = $1, updatedAt = CURRENT_TIMESTAMP
		WHERE username = $2 AND chatReference = $3
		RETURNING id, username, chatReference, status, createdAt, updatedAt`,
		status, username, chatReference,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&participant.Id, &participant.Username, &participant.ChatReference,
			&participant.Status, &participant.CreatedAt, &participant.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return participant, nil
	}
	return nil, errors.New("participant not found")
}

func HasPendingParticipant(chatReference string) (bool, error) {
	var count int
	row := Instance.QueryRow(
		`SELECT COUNT(*) FROM participants WHERE chatReference = $1 AND status = $2`,
		chatReference, ParticipantStatusPending,
	)
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func DeleteParticipant(username string, chatReference string) (*Participant, error) {
	participant := &Participant{}
	rows, err := Instance.Query(
		`DELETE FROM participants WHERE username = $1 AND chatReference = $2
		RETURNING id, username, chatReference, status, createdAt, updatedAt`,
		username, chatReference,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(
			&participant.Id, &participant.Username, &participant.ChatReference,
			&participant.Status, &participant.CreatedAt, &participant.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return participant, nil
	}
	return nil, nil
}

func GetChatRefFor(user string, other string) (*string, error) {
	var ref string
	rows, err := Instance.Query(
		`SELECT chatReference FROM participants
		WHERE username IN ($1, $2) GROUP BY chatReference HAVING COUNT(chatReference) = 2`,
		user, other,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	hasRows := rows.Next()
	if hasRows {
		scanErr := rows.Scan(&ref)
		if scanErr != nil {
			println("failed due to scan error")
			return nil, scanErr
		}

		if len(ref) == 0 {
			return nil, errors.New("reference can't be empty")
		}
		println("shared ref: ", ref)
		return &ref, nil
	}
	return nil, nil
}
