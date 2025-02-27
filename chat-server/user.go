package chatserver

import (
	"github.com/te6lim/go-chat/tracer"
)

type User struct {
	Message           chan Message
	Username          string
	Followers         []string
	Session           chan *UserSession
	PrivateRooms      map[string]*TwoUserRoom
	RequestToJoinRoom chan string
	Tracer            tracer.Tracer
}

func (user *User) ListenForNewRoom() {
	for session := range user.Session {
		user.PrivateRooms[session.withUser] = session.room
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
