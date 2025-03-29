package database

import "errors"

type Participant struct {
	Id            string `json:"id"`
	Username      string `json:"username"`
	ChatReference string `json:"chatReference"`
	CreatedAt     string `json:"createdAt"`
}

func InsertParticipant(participant Participant) (*Participant, error) {
	newParticipant := &Participant{}
	if len(participant.Username) == 0 || len(participant.ChatReference) == 0 {
		return nil, errors.New("invalid participant")
	}
	err := Instance.QueryRow(
		`INSERT INTO participants (username, chatReference) VALUES ($1, $2) RETURNING id, username, chatReference, createdAt`,
		participant.Username, participant.ChatReference,
	).Scan(&newParticipant.Id, &newParticipant.Username, &newParticipant.ChatReference, &newParticipant.CreatedAt)

	if err != nil {
		newParticipant = nil
	}
	return newParticipant, nil
}

func GetParticipantsInChat(chatReference string) []Participant {
	participants := []Participant{}
	rows, err := Instance.Query(
		`SELECT id, chatReference, createdAt FROM participants WHERE chatReference = $1`,
		chatReference,
	)
	for rows.Next() {
		participant := &Participant{}
		rows.Scan(&participant.Id, &participant.Username, &participant.ChatReference, &participant.CreatedAt)
		participants = append(participants, *participant)
	}
	if err != nil {
		participants = nil
	}
	return participants
}

func GetParticipant(username string, chatReference string) *Participant {
	participant := &Participant{}
	err := Instance.QueryRow(
		`SELECT id, username, chatReference, createdAt FROM participants WHERE username = $1 AND chatReference = $2`,
		username, chatReference,
	).Scan(&participant.Id, &participant.Username, &participant.ChatReference, &participant.CreatedAt)
	if err != nil {
		participant = nil
	}
	return participant
}

func GetChatRefFor(user string, other string) *string {
	var ref *string
	err := Instance.QueryRow(
		`SELECT chatReference FROM participants WHERE username IN ($1, $2) GROUP BY chatReference HAVING COUNT(chatReference) = 2`,
		user, other,
	).Scan(&ref)
	if err != nil {
		ref = nil
	}
	return ref
}
