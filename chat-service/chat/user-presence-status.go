package chat

type PresenceStatus int

const (
	ONLINE PresenceStatus = iota
	AWAY
	OFFLINE
)

func (activity PresenceStatus) GetStatus() string {
	switch activity {
	case ONLINE:
		return "ONLINE"
	case AWAY:
		return "AWAY"
	case OFFLINE:
		return "OFFLINE"
	}
	return ""
}
