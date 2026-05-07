package database

type RevokedInvite struct {
	Id                string `json:"id"`
	InviteeUsername   string `json:"inviteeUsername"`
	InitiatorUsername string `json:"initiatorUsername"`
	ChatReference     string `json:"chatReference"`
	RevokedAt         string `json:"revokedAt"`
}

func StoreRevokedInvite(inviteeUsername, initiatorUsername, chatRef string) error {
	_, err := Instance.Exec(
		`INSERT INTO revoked_invites (inviteeUsername, initiatorUsername, chatReference)
		 VALUES ($1, $2, $3) ON CONFLICT (inviteeUsername, chatReference) DO NOTHING`,
		inviteeUsername, initiatorUsername, chatRef,
	)
	return err
}

func GetRevokedInvitesForUser(username string) ([]RevokedInvite, error) {
	rows, err := Instance.Query(
		`SELECT id, inviteeUsername, initiatorUsername, chatReference, revokedAt
		 FROM revoked_invites WHERE inviteeUsername = $1`,
		username,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []RevokedInvite
	for rows.Next() {
		r := RevokedInvite{}
		if scanErr := rows.Scan(&r.Id, &r.InviteeUsername, &r.InitiatorUsername, &r.ChatReference, &r.RevokedAt); scanErr != nil {
			return nil, scanErr
		}
		result = append(result, r)
	}
	return result, nil
}

func DeleteRevokedInvite(inviteeUsername, chatRef string) error {
	_, err := Instance.Exec(
		`DELETE FROM revoked_invites WHERE inviteeUsername = $1 AND chatReference = $2`,
		inviteeUsername, chatRef,
	)
	return err
}
