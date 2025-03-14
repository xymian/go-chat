package chat

import (
	"errors"

	"github.com/te6lim/go-chat/tracer"
)

var OnlineUsers = make(map[string]*Socketuser)
var NewUser chan *Socketuser = make(chan *Socketuser)
var LoggedOutUser chan *Socketuser = make(chan *Socketuser)

type UserListeners struct {
	SendMessage       chan SocketMessage
	ReceiveMessage    chan SocketMessage
	ChatSession       chan *ChatSession
	RequestToJoinRoom chan string
}

type Socketuser struct {
	UserListeners
	Username        string
	PrivateSessions map[string]*ChatSession
	Tracer          tracer.Tracer
}

func CreateNewUser(username string) *Socketuser {
	return &Socketuser{
		Username:        username,
		PrivateSessions: make(map[string]*ChatSession),
		Tracer:          tracer.New(),

		UserListeners: UserListeners{
			SendMessage:       make(chan SocketMessage),
			ReceiveMessage:    make(chan SocketMessage),
			ChatSession:       make(chan *ChatSession),
			RequestToJoinRoom: make(chan string),
		},
	}
}

func (session *ChatSession) ForwardMessageToRoomMembers(message SocketMessage) {
	OnlineUsers[session.User].PrivateSessions[session.OtherUser].Room.ForwardedMessage <- message
}

func (user *Socketuser) LeaveRoom(otherUser string) {
	user.PrivateSessions[otherUser].Room.leave <- user
}

func (user *Socketuser) JoinRoomWith(otherUser string) error {
	room := user.PrivateSessions[otherUser].Room
	if len(room.participants) < 2 {
		room.join <- user
		return nil
	}
	return errors.New("room is full. please create another room with this user")
}

func (user *Socketuser) ListenForNewChatSession() {
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

func (user *Socketuser) ListenForJoinRoomRequest() {
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
