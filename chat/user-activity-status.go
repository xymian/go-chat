package chat

type Activity int

const (
	ONLINE Activity = iota
	AWAY
	OFFLINE
)

func (activity Activity) GetStatus() string {
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
