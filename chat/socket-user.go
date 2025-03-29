package chat

import (
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/tracer"
)

var OnlineUsers = make(map[string]*Socketuser)
var NewUser chan *Socketuser = make(chan *Socketuser)
var LoggedOutUser chan *Socketuser = make(chan *Socketuser)

var AskForUserToChatWith = make(chan *Socketuser)

type UserListeners struct {
	Conn           *websocket.Conn
	SendMessage    chan database.Message
	ReceiveMessage chan database.Message
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
	var room *Room
	if Rooms[chatReference] == nil {
		room = CreateRoom(chatReference)
		AddRoom <- room
		go room.Run()
	} else {
		room = Rooms[chatReference]
	}
	endpoint := fmt.Sprintf("/room/%s", chatReference)
	config.Router.Handle(endpoint, room)
}

func CreateNewUser(username string) *Socketuser {
	return &Socketuser{
		Username:   username,
		SessionIds: make(map[string]bool),
		Tracer:     tracer.New(),

		UserListeners: UserListeners{
			SendMessage:    make(chan database.Message),
			ReceiveMessage: make(chan database.Message),
			Room:           make(chan *Room),
		},
	}
}

func (user *Socketuser) ReadMessages(room *Room) {
	defer func() {
		user.Conn.Close()
		user.Tracer.Trace("connection closed")
	}()
	for {
		var newMessage *database.Message
		err := user.Conn.ReadJSON(&newMessage)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
		room.ForwardedMessage <- *newMessage
	}
}

func (user *Socketuser) MessageReceiver(room *Room) {
	defer func() {
		user.Tracer.Trace("done receiving")
	}()
	for message := range user.ReceiveMessage {
		if room.participants[user] {
			user.Conn.WriteJSON(message)
			user.Tracer.Trace("message: ", message.TextMessage, "from", message.SenderUsername, "has been received")
		} else {
			user.Tracer.Trace("You are not in this room")
		}
	}
}

func (user *Socketuser) LeaveRoom(room *Room) {
	room.leave <- user
}

func (room *Room) JoinRoom(user *Socketuser) error {
	if !room.participants[user] {
		if len(room.participants) < 2 {
			room.join <- user
			return nil
		} else {
			return errors.New("room is full. please create another room with this user")
		}
	} else {
		return errors.New("user is already in the room")
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
