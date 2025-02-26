package chatserver

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/tracer"
)

type User struct {
	Message           chan Message
	Username          string
	Followers         []string
	NewRoom           chan *TwoUserRoomPayload
	PrivateRooms      map[string]*TwoUserRoom
	RequestToJoinRoom chan string
	Tracer            *tracer.EventTracer
}

func (user *User) ListenForNewRoom() {
	for newRoom := range user.NewRoom {
		user.PrivateRooms[newRoom.Id] = newRoom.Room
		newRoom.Room.Tracer.Trace("New room added")
	}
}

var OnlineUsers = make(map[string]*User)
var NewUser chan *User = make(chan *User)
var LoggedOutUser chan *User = make(chan *User)

func (user *User) ReadFromConnectionWith(otherUser *User) {
	privateRoom := user.PrivateRooms[otherUser.Username]
	defer privateRoom.Conn.Close()
	for {
		_, message, err := privateRoom.Conn.ReadMessage()
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
		user.PrivateRooms[otherUser.Username].Forward(Message{Text: message, Sender: ""})
	}
}

func (user *User) WriteToConnectionWith(otherUser *User) {
	privateRoom := user.PrivateRooms[otherUser.Username]
	defer privateRoom.Conn.Close()
	for m := range user.Message {
		err := privateRoom.Conn.WriteMessage(websocket.TextMessage, m.Text)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
	}
}

func ListenForActiveUsers() {
	for {
		select {
		case newUser := <-NewUser:
			OnlineUsers[newUser.Username] = newUser
			newUser.Tracer.Trace("New user ", newUser.Username, " is online")

		case loggedOutUser := <-LoggedOutUser:
			OnlineUsers[loggedOutUser.Username] = nil
			loggedOutUser.Tracer.Trace("User ", loggedOutUser.Username, " logged out")
		}
	}
}

func (user *User) ListenForJoinRoomRequest() {
	for roomWith := range user.RequestToJoinRoom {
		if user.PrivateRooms[roomWith] != nil {
			user.PrivateRooms[roomWith].Join(user)
			user.Tracer.Trace("User joined ", roomWith, " asked to join the room")
		}
	}
}
