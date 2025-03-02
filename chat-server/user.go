package chatserver

import (
	"github.com/te6lim/go-chat/tracer"
)

type User struct {
	Message           chan Message
	Username          string
	Followers         []string
	Session           chan *UserSession
	PrivateRooms      map[string]*twoUserRoom
	RequestToJoinRoom chan string
	Tracer            tracer.Tracer
}

func CreateNewUser(username string) *User {
	return &User{
		Message:           make(chan Message),
		Username:          username,
		Session:           make(chan *UserSession),
		PrivateRooms:      make(map[string]*twoUserRoom),
		RequestToJoinRoom: make(chan string),
		Tracer:            tracer.New(),
	}
}

func (user *User) ListenForNewRoom() {
	for session := range user.Session {
		user.PrivateRooms[session.WithUser] = session.Room
		session.Room.Tracer.Trace(user.Username, " is in session with ", session.WithUser)
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
