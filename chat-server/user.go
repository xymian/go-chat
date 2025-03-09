package chatserver

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/tracer"
)

type User struct {
	Conn              *websocket.Conn
	SendMessage       chan Message
	ReceiveMessage    chan Message
	Username          string
	Followers         []string
	Session           chan *UserSession
	PrivateRooms      map[string]*room
	RequestToJoinRoom chan string
	Tracer            tracer.Tracer
}

func CreateNewUser(username string) *User {
	return &User{
		SendMessage:       make(chan Message),
		ReceiveMessage:    make(chan Message),
		Username:          username,
		Session:           make(chan *UserSession),
		PrivateRooms:      make(map[string]*room),
		RequestToJoinRoom: make(chan string),
		Tracer:            tracer.New(),
	}
}

func (user *User) ListenForNewRoom() {
	for session := range user.Session {
		user.PrivateRooms[session.OtherUser] = session.Room
		session.Room.Tracer.Trace(user.Username, " is in session with ", session.OtherUser)
	}
}

var OnlineUsers = make(map[string]*User)
var NewUser chan *User = make(chan *User)
var LoggedOutUser chan *User = make(chan *User)

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

func (user *User) ListenForJoinRoomRequest() {
	for from := range user.RequestToJoinRoom {
		if user.PrivateRooms[from] == nil {
			requestingUser := OnlineUsers[from]
			if requestingUser != nil {
				user.Tracer.Trace(from, " is requesting to chat with ", user.Username)
				CreateSession(user, requestingUser.Username)
			}
		}
	}
}

func (user *User) MessageSender() {
	defer func() {
		user.Conn.Close()
		user.Tracer.Trace("connection closed")
	}()
	for message := range user.SendMessage {
		err := user.Conn.WriteJSON(message)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
	}
}

func (user *User) MessageReceiver() {
	defer func() {
		user.Conn.Close()
		user.Tracer.Trace("connection closed")
	}()
	for message := range user.ReceiveMessage {
		user.Tracer.Trace("message: ", message, "has been received")
		// save message to db or something
	}
}
