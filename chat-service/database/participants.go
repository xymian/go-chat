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

// GetPendingInvitesForUser returns all chats where the given user has a PENDING
// participant status, along with the initiator's username for each invite.
func GetPendingInvitesForUser(username string) ([]Participant, error) {
	rows, err := Instance.Query(
		`SELECT id, username, chatReference, status, createdAt, updatedAt
		FROM participants WHERE username = $1 AND status = $2`,
		username, ParticipantStatusPending,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []Participant
	for rows.Next() {
		p := Participant{}
		if scanErr := rows.Scan(&p.Id, &p.Username, &p.ChatReference, &p.Status, &p.CreatedAt, &p.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}
		participants = append(participants, p)
	}
	return participants, nil
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

// ConversationSummary holds the data the Android client needs to seed its local DB.
type ConversationSummary struct {
	ChatReference string `json:"chatReference"`
	OtherUsername string `json:"otherUsername"`
	ChatType      string `json:"chatType"`
}

// GetConversationsForUser returns one summary per accepted private chat the user
// belongs to. For group chats it returns the chat name as OtherUsername.
func GetConversationsForUser(username string) ([]ConversationSummary, error) {
	rows, err := Instance.Query(
		`SELECT p2.chatReference, COALESCE(p2.username, c.name, ''), c.chatType
		 FROM participants p1
		 JOIN participants p2
		   ON p1.chatReference = p2.chatReference AND p2.username != p1.username
		 JOIN chats c ON c.chatReference = p1.chatReference
		 WHERE p1.username = $1
		   AND p1.status = $2
		   AND p2.status = $2`,
		username, ParticipantStatusAccepted,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ConversationSummary
	for rows.Next() {
		s := ConversationSummary{}
		if scanErr := rows.Scan(&s.ChatReference, &s.OtherUsername, &s.ChatType); scanErr != nil {
			return nil, scanErr
		}
		results = append(results, s)
	}
	return results, nil
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
