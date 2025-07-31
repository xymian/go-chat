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
	rows, err := Instance.Query(
		`INSERT INTO participants (username, chatReference) VALUES ($1, $2) RETURNING id, username, chatReference, createdAt`,
		participant.Username, participant.ChatReference,
	)

	if err != nil {
		return nil, nil
	}

	rows.Next()
	scanErr := rows.Scan(&newParticipant.Id, &newParticipant.Username, &newParticipant.ChatReference, &newParticipant.CreatedAt)
	if scanErr != nil {
		return nil, scanErr
	}
	return newParticipant, nil
}

func GetParticipantsInChat(chatReference string) ([]Participant, error) {
	participants := []Participant{}
	rows, err := Instance.Query(
		`SELECT id, username, chatReference, createdAt FROM participants WHERE chatReference = $1`,
		chatReference,
	)
	if err != nil {
		return []Participant{}, nil
	}

	for rows.Next() {
		participant := &Participant{}
		scanErr := rows.Scan(&participant.Id, &participant.Username, &participant.ChatReference, &participant.CreatedAt)
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
		`SELECT id, username, chatReference, createdAt FROM participants WHERE username = $1 AND chatReference = $2`,
		username, chatReference,
	)

	if err != nil {
		return nil, nil
	}

	rows.Next()
	scanErr := rows.Scan(&participant.Id, &participant.Username, &participant.ChatReference, &participant.CreatedAt)
	if scanErr != nil {
		return nil, scanErr
	}
	return participant, nil
}

func GetChatRefFor(user string, other string) (*string, error) {
	var ref string
	rows, err := Instance.Query(
		`SELECT chatReference FROM participants WHERE username IN ($1, $2) GROUP BY chatReference HAVING COUNT(chatReference) = 2`,
		user, other,
	)
	if err != nil {
		return nil, nil
	}

	scanErr := rows.Scan(&ref)
	if scanErr != nil {
		return nil, scanErr
	}

	if len(ref) == 0 {
		return nil, errors.New("reference can't be empty")
	}
	return &ref, nil
}
