package chat

import (
	"github.com/te6lim/go-chat/tracer"
)

var OnlineUsers = make(map[string]*Socketuser)
var NewUser chan *Socketuser = make(chan *Socketuser)
var LoggedOutUser chan *Socketuser = make(chan *Socketuser)

var AskForUserToChatWith = make(chan *Socketuser)

type UserListeners struct {
	SendMessage       chan SocketMessage
	ReceiveMessage    chan SocketMessage
	Room              chan *Room
}

type JoinSessionRequest struct {
	RoomId      string
	RequestingUser string
}

type Socketuser struct {
	UserListeners
	Username   string
	SessionIds map[string]bool
	Tracer     tracer.Tracer
}

func CreateNewUser(username string) *Socketuser {
	return &Socketuser{
		Username:   username,
		SessionIds: make(map[string]bool),
		Tracer:     tracer.New(),

		UserListeners: UserListeners{
			SendMessage:       make(chan SocketMessage),
			ReceiveMessage:    make(chan SocketMessage),
			Room:              make(chan *Room),
		},
	}
}

func ListenForActiveUsers() {
	for {
		select {
		case newUser := <-NewUser:
			OnlineUsers[newUser.Username] = newUser
			newUser.Tracer.Trace("\nNew User", newUser.Username, " is online")

		case loggedOutUser := <-LoggedOutUser:
			OnlineUsers[loggedOutUser.Username] = nil
			loggedOutUser.Tracer.Trace("User", loggedOutUser.Username, " logged out")
		}
	}
}
