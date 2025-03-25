package chat

import (
	"fmt"

	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/tracer"
)

var OnlineUsers = make(map[string]*Socketuser)
var NewUser chan *Socketuser = make(chan *Socketuser)
var LoggedOutUser chan *Socketuser = make(chan *Socketuser)

var AskForUserToChatWith = make(chan *Socketuser)

type UserListeners struct {
	SendMessage    chan SocketMessage
	ReceiveMessage chan SocketMessage
	Room           chan *Room
}

type JoinSessionRequest struct {
	RoomId         string
	RequestingUser string
}

type Socketuser struct {
	UserListeners
	Username   string
	SessionIds map[string]bool
	Tracer     tracer.Tracer
}

func SetupSocketUser(username string, otherUsername string, chatReference string) {
	var newUser *Socketuser
	if OnlineUsers[username] != nil {
		newUser = OnlineUsers[username]
		newUser.Tracer.Trace("\nUser", username, " is online")
	} else {
		newUser = CreateNewUser(username)
		NewUser <- newUser
	}

	var room *Room
	if Rooms[chatReference] == nil {
		room = CreateRoom(chatReference)
		AddRoom <- room
		go room.Run()
	} else {
		room = Rooms[chatReference]
	}

	endpoint := fmt.Sprintf("/chat/%s", room.Id)
	config.Router.Handle(endpoint, room)
	room.JoinRoom(newUser)
	//go room.MessageSender(user)
	go room.MessageReceiver(newUser)
}

func CreateNewUser(username string) *Socketuser {
	return &Socketuser{
		Username:   username,
		SessionIds: make(map[string]bool),
		Tracer:     tracer.New(),

		UserListeners: UserListeners{
			SendMessage:    make(chan SocketMessage),
			ReceiveMessage: make(chan SocketMessage),
			Room:           make(chan *Room),
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
