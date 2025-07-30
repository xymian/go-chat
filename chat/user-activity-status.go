package chat

type Activity int

const (
	ONLINE Activity = iota
	AWAY
	OFFLINE
)