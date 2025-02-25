package chatserver

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type User struct {
	Message           chan Message
	Username          string
	Followers         []string
	NewRoom           chan *TwoUserRoomPayload
	PrivateRooms      map[string]*TwoUserRoom
	RequestToJoinRoom chan string
}

func (user *User) ListenForNewRoom() {
	for newRoom := range NewRoom {
		if user.NewRoom != nil {
			user.PrivateRooms[newRoom.Id] = newRoom.Room
		}
	}
}

var OnlineUsers = make(map[string]*User)
var NewUser chan *User
var LoggedOutUser chan *User

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

		case loggedOutUser := <-LoggedOutUser:
			OnlineUsers[loggedOutUser.Username] = nil
		}
	}
}

func (user *User) ListenForJoinRoomRequest() {
	for roomWith := range user.RequestToJoinRoom {
		if user.PrivateRooms[roomWith] != nil {
			user.PrivateRooms[roomWith].Join(user)
		}
	}
}
