package requests

type AckRequest struct {
	From          string `json:"from"`
	To            string `json:"to"`
	ChatReference string `json:"chatReference"`
	Username      string `json:"username"`
}
