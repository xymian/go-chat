package responses

type UserChatInfo struct {
	Username        string `json:"username"`
	DisplayImageUrl string `json:"displayImageUrl"`
	ChatReference   string `json:"chatReference"`
}
