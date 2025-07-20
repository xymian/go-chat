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
		return nil, err
	}
	return newParticipant, nil
}

func GetParticipantsInChat(chatReference string) ([]Participant, error) {
	participants := []Participant{}
	rows, err := Instance.Query(
		`SELECT id, username, chatReference, createdAt FROM participants WHERE chatReference = $1`,
		chatReference,
	)
	for rows.Next() {
		participant := &Participant{}
		scanErr := rows.Scan(&participant.Id, &participant.Username, &participant.ChatReference, &participant.CreatedAt)
		if (scanErr != nil) {
			return []Participant{}, scanErr
		}
		participants = append(participants, *participant)
	}
	if err != nil {
		return []Participant{}, err
	}
	return participants, nil
}

func GetParticipant(username string, chatReference string) (*Participant, error) {
	participant := &Participant{}
	err := Instance.QueryRow(
		`SELECT id, username, chatReference, createdAt FROM participants WHERE username = $1 AND chatReference = $2`,
		username, chatReference,
	).Scan(&participant.Id, &participant.Username, &participant.ChatReference, &participant.CreatedAt)
	if err != nil {
		return nil, err
	}
	return participant, nil
}

func GetChatRefFor(user string, other string) (*string, error) {
	var ref string
	err := Instance.QueryRow(
		`SELECT chatReference FROM participants WHERE username IN ($1, $2) GROUP BY chatReference HAVING COUNT(chatReference) = 2`,
		user, other,
	).Scan(&ref)
	if err != nil {
		return nil, err
	}
	if len(ref) == 0 {
		return nil, errors.New("reference can't be empty")
	}
	return &ref, nil
}
