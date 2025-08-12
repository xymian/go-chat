package chat

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/tracer"
)

var OnlineUsers = make(map[string]*Socketuser)
var NewUser chan *Socketuser = make(chan *Socketuser)
var LoggedOutUser chan *Socketuser = make(chan *Socketuser)

var AskForUserToChatWith = make(chan *Socketuser)

type PrivateChat struct {
	PrivateConn    *websocket.Conn
	SendMessage    chan database.Message
	ReceiveMessage chan database.Message
}

type Socketuser struct {
	PrivateChat
	*Conversations
	Username string
	Activity Activity
	Tracer   tracer.Tracer
}

func CreateNewSocketUser(user *database.User, activity Activity) (*Socketuser, error) {
	conversations := &Conversations{
		Chats: map[string]bool{},
	}
	conversations.Tracer = tracer.New()
	conversationMap := map[int64]string{}
	err := json.Unmarshal([]byte(*user.Interactions), &conversationMap)
	if err != nil {
		return nil, err
	}
	for _, chatRef := range conversationMap {
		conversations.Chats[chatRef] = true
	}
	return &Socketuser{
		Username: user.Username,
		Activity: activity,
		Tracer:   tracer.New(),

		PrivateChat: PrivateChat{
			SendMessage:    make(chan database.Message),
			ReceiveMessage: make(chan database.Message),
		},

		Conversations: conversations,
	}, nil
}

func ListenForActiveUsers() {
	for {
		select {
		case newUser := <-NewUser:
			OnlineUsers[newUser.Username] = newUser
			newUser.Tracer.Trace("number of users: ", len(OnlineUsers))
			newUser.Tracer.Trace("New User", newUser.Username, " is ", newUser.Activity.GetStatus())

		case loggedOutUser := <-LoggedOutUser:
			delete(OnlineUsers, loggedOutUser.Username)
			loggedOutUser.Tracer.Trace("User", loggedOutUser.Username, " logged out")
		}
	}
}
