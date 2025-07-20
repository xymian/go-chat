package models

type DeliverMessage struct {
	Sender         string            `json:"sender"`
	MessageDetails map[string]string `json:"messageDetails"`
	ChatReference  string            `json:"chatReference"`
	Seen           bool              `json:"seen"`
}

type AckRequest struct {
	From          string `json:"from"`
	To            string `json:"to"`
	ChatReference string `json:"chatReference"`
	Username      string `json:"username"`
}
