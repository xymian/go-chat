package models


type Response [T any] struct {
	Data *T `json:"data"`
	Message string `json:"message"`
	Error string `json:"error"`
	StatusCode int `json:"statusCode"`
	IsSuccessful bool `json:"isSuccessful"`
}

type NewChat struct {
	User          string `json:"user"`
	Other         string `json:"other"`
	ChatReference string `json:"chatReference"`
}

type UserChatInfo struct {
	Username        string `json:"username"`
	DisplayImageUrl string `json:"displayImageUrl"`
	ChatReference   string `json:"chatReference"`
}

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