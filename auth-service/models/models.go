package models


type Response [T any] struct {
	Data *T `json:"data"`
	Message string `json:"message"`
	Error string `json:"error"`
	StatusCode int `json:"statusCode"`
	IsSuccessful bool `json:"isSuccessful"`
}