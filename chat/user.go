package chat

import (
	"fmt"

	"github.com/te6lim/go-chat/tracer"
)

var OnlineUsers = make(map[string]*User)
var NewUser chan *User = make(chan *User)
var LoggedOutUser chan *User = make(chan *User)

type UserListeners struct {
	SendMessage       chan Message
	ReceiveMessage    chan Message
	ChatSession       chan *ChatSession
	RequestToJoinRoom chan string
}

type User struct {
	UserListeners
	Username        string
	PrivateSessions map[string]*ChatSession
	Tracer          tracer.Tracer
}

func CreateNewUser(username string) *User {
	return &User{
		Username:        username,
		PrivateSessions: make(map[string]*ChatSession),
		Tracer:          tracer.New(),

		UserListeners: UserListeners{
			SendMessage:       make(chan Message),
			ReceiveMessage:    make(chan Message),
			ChatSession:       make(chan *ChatSession),
			RequestToJoinRoom: make(chan string),
		},
	}
}

func (user *User) ListenForNewRoom() {
	for session := range user.ChatSession {
		user.PrivateSessions[session.OtherUser] = session
		session.Room.Tracer.Trace(user.Username, " is in session with ", session.OtherUser)
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

func (user *User) ListenForJoinRoomRequest() {
	for from := range user.RequestToJoinRoom {
		if user.PrivateSessions[from] == nil {
			requestingUser := OnlineUsers[from]
			if requestingUser != nil {
				user.Tracer.Trace(from, " is requesting to chat with ", user.Username)
				CreateSession(user, requestingUser.Username)
			}
		}
	}
}

func (session *ChatSession) MessageSender() {
	user := OnlineUsers[session.User]
	defer func() {
		session.SharedClientConnection.Close()
		user.Tracer.Trace("connection closed")
	}()
	for message := range user.SendMessage {
		err := session.SharedClientConnection.WriteJSON(message)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
	}
}

func (session *ChatSession) MessageReceiver() {
	user := OnlineUsers[session.User]
	defer func() {
		session.SharedClientConnection.Close()
		user.Tracer.Trace("connection closed")
	}()
	for message := range user.ReceiveMessage {
		user.Tracer.Trace("message: ", message.Text, "from", message.Sender, "has been received")
		// save message to db or something
	}
}
