package responses

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