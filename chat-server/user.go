package chatserver

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/tracer"
)

type User struct {
	Connections       map[string]*websocket.Conn
	SendMessage       chan Message
	ReceiveMessage    chan Message
	Username          string
	Followers         []string
	Session           chan *ChatSession
	PrivateRooms      map[string]*room
	RequestToJoinRoom chan string
	Tracer            tracer.Tracer
}

func CreateNewUser(username string) *User {
	return &User{
		Connections:       make(map[string]*websocket.Conn),
		SendMessage:       make(chan Message),
		ReceiveMessage:    make(chan Message),
		Username:          username,
		Session:           make(chan *ChatSession),
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

func (session *ChatSession) MessageSender() {
	defer func() {
		session.User.Connections[session.OtherUser].Close()
		session.User.Tracer.Trace("connection closed")
	}()
	for message := range session.User.SendMessage {
		err := session.User.Connections[session.OtherUser].WriteJSON(message)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
	}
}

func (session *ChatSession) MessageReceiver() {
	defer func() {
		session.User.Connections[session.OtherUser].Close()
		session.User.Tracer.Trace("connection closed")
	}()
	for message := range session.User.ReceiveMessage {
		session.User.Tracer.Trace("message: ", message.Text, "from", message.Sender, "has been received")
		// save message to db or something
	}
}
