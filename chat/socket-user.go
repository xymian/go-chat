package chat

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/tracer"
)

var ActiveSocketUsers = make(map[string]*Socketuser)
var NewUserFromRoomSetup chan *Socketuser = make(chan *Socketuser)
var NewUserFromConversationsSetup chan *Socketuser = make(chan *Socketuser)
var LoggedOutUser chan *Socketuser = make(chan *Socketuser)

type PrivateChat struct {
	PrivateConn    *websocket.Conn
	ReceiveMessage chan database.Message
}

type Socketuser struct {
	PrivateChat
	*Conversations
	Username string
	Activity Activity
	Tracer   tracer.Tracer
}

func SetupSocketUser(user *database.User, activity Activity) (*Socketuser, error) {
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
			ReceiveMessage: make(chan database.Message),
		},

		Conversations: conversations,
	}, nil
}

func ListenForActiveUsers() {
	for {
		select {
		case newUser := <-NewUserFromRoomSetup:
			activeUser := ActiveSocketUsers[newUser.Username]
			if activeUser != nil {
				activeUser.PrivateConn = newUser.PrivateConn
				activeUser.ReceiveMessage = newUser.ReceiveMessage
				activeUser.Activity = newUser.Activity
			} else {
				ActiveSocketUsers[newUser.Username] = newUser
				activeUser = newUser
			}

			newUser.Tracer.Trace("number of users: ", len(ActiveSocketUsers))
			newUser.Tracer.Trace("New User", newUser.Username, " is ", activeUser.Activity.GetStatus())

		case newUser := <-NewUserFromConversationsSetup:
			activeUser := ActiveSocketUsers[newUser.Username]
			if activeUser != nil {
				activeUser.PublicConn = newUser.PublicConn
				activeUser.IReceiveMessage = newUser.IReceiveMessage
				activeUser.Chats = newUser.Chats
				activeUser.Activity = newUser.Activity
			} else {
				ActiveSocketUsers[newUser.Username] = newUser
				activeUser = newUser
			}
			newUser.Tracer.Trace("number of users: ", len(ActiveSocketUsers))
			newUser.Tracer.Trace("New User", newUser.Username, " is ", activeUser.Activity.GetStatus())

		case loggedOutUser := <-LoggedOutUser:
			delete(ActiveSocketUsers, loggedOutUser.Username)
			loggedOutUser.Tracer.Trace("User", loggedOutUser.Username, " logged out")
		}
	}
}
