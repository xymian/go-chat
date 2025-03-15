package database

type Participant struct {
	Id        string `json:"id"`
	UserId    string `json:"userId"`
	ChatId    string `json:"chatId"`
	CreatedAt string `json:"createdAt"`
}
