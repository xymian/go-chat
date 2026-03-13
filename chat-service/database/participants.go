package database

import (
	"errors"
	"log"
)

const RoleAdmin = "admin"
const RoleMember = "member"

type Participant struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	ChatReference string `json:"chatReference"`
	Role          string `json:"role"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

func InsertParticipant(participant Participant) (*Participant, error) {
	if participant.Role == "" {
		participant.Role = RoleMember
	}
	newParticipant := &Participant{}
	if len(participant.Username) == 0 || len(participant.ChatReference) == 0 {
		return nil, errors.New("invalid participant")
	}
	rows, err := Instance.Query(
		`INSERT INTO participants (username, chatReference, role)
		VALUES ($1, $2, $3)
		RETURNING id, username, chatReference, role, createdAt, updatedAt`,
		participant.Username, participant.ChatReference, participant.Role,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&newParticipant.Id, &newParticipant.Username, &newParticipant.ChatReference,
			&newParticipant.Role, &newParticipant.CreatedAt, &newParticipant.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return newParticipant, nil
	}
	return nil, nil
}

func RemoveParticipant(username string, chatReference string) (*Participant, error) {
	p := &Participant{}
	rows, err := Instance.Query(
		`DELETE FROM participants WHERE username = $1 AND chatReference = $2
		RETURNING id, username, chatReference, role, createdAt, updatedAt`,
		username, chatReference,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&p.Id, &p.Username, &p.ChatReference,
			&p.Role, &p.CreatedAt, &p.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return p, nil
	}
	return nil, nil
}

func GetParticipantsInChat(chatReference string) ([]Participant, error) {
	participants := []Participant{}
	rows, err := Instance.Query(
		`SELECT id, username, chatReference, role, createdAt, updatedAt
		FROM participants WHERE chatReference = $1`,
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
			&participant.Role, &participant.CreatedAt, &participant.UpdatedAt,
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
		`SELECT id, username, chatReference, role, createdAt, updatedAt
		FROM participants WHERE username = $1 AND chatReference = $2`,
		username, chatReference,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	if rows.Next() {
		scanErr := rows.Scan(
			&participant.Id, &participant.Username, &participant.ChatReference,
			&participant.Role, &participant.CreatedAt, &participant.UpdatedAt,
		)
		if scanErr != nil {
			return nil, scanErr
		}
		return participant, nil
	}
	return nil, nil
}

func IsAdmin(username string, chatReference string) (bool, error) {
	p, err := GetParticipant(username, chatReference)
	if err != nil {
		return false, err
	}
	if p == nil {
		return false, nil
	}
	return p.Role == RoleAdmin, nil
}

// GetChatRefsForUser returns all chat references the user participates in.
func GetChatRefsForUser(username string) ([]string, error) {
	refs := []string{}
	rows, err := Instance.Query(
		`SELECT chatReference FROM participants WHERE username = $1`,
		username,
	)
	if err != nil {
		return refs, err
	}
	defer rows.Close()

	for rows.Next() {
		var ref string
		if scanErr := rows.Scan(&ref); scanErr != nil {
			return refs, scanErr
		}
		refs = append(refs, ref)
	}
	return refs, nil
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

	if rows.Next() {
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
