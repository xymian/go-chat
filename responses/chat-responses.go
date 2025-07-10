package responses

type NewChat struct {
	User          string `json:"user"`
	Other         string `json:"other"`
	ChatReference string `json:"chatReference"`
}
