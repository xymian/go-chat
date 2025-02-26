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
	}
}

var OnlineUsers = make(map[string]*User)
var NewUser chan *User = make(chan *User)
var LoggedOutUser chan *User = make(chan *User)

func (user *User) ReadFromConnectionWith(otherUser string) {
	privateRoom := user.PrivateRooms[otherUser]
	defer func() {
		if privateRoom != nil {
			defer privateRoom.Conn.Close()
		}
	}()
	for {
		if privateRoom != nil {
			_, message, err := privateRoom.Conn.ReadMessage()
			if err != nil {
				fmt.Println("Connection error: ", err)
				return
			}
			privateRoom.Forward(Message{Text: message, Sender: ""})
		}
	}
}

func (user *User) WriteToConnectionWith(otherUser string) {
	privateRoom := user.PrivateRooms[otherUser]
	defer func() {
		if privateRoom != nil {
			privateRoom.Conn.Close()
		}
	}()

	for m := range user.Message {
		if privateRoom != nil {
			err := privateRoom.Conn.WriteMessage(websocket.TextMessage, m.Text)
			if err != nil {
				fmt.Println("Connection error: ", err)
				return
			}
		}
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
	for roomWith := range user.RequestToJoinRoom {
		if user.PrivateRooms[roomWith] != nil {
			user.PrivateRooms[roomWith].Join(user)
			user.Tracer.Trace("User joined ", roomWith, " asked to join the room")
		}
	}
}
