package models

type DeliverMessages struct {
	Sender          string           `json:"sender"`
	MessagesDetails []MessageDetails `json:"messagesDetails"`
	ChatReference   string           `json:"chatReference"`
}

type MessageDetails struct {
	MessageReference     string `json:"messageReference"`
	SentTimestamp        string `json:"sentTimestamp"`
	DeliveredTimestamp   string `json:"deliveredTimestamp"`
	ReadTimestamp        string `json:"readTimestamp"`
	IsReadReceiptEnabled string `json:"isReadReceiptEnabled"`
}

type AckRequest struct {
	From          string `json:"from"`
	To            string `json:"to"`
	ChatReference string `json:"chatReference"`
	Username      string `json:"username"`
}
