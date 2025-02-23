package chatserver

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/socket"
)

type User struct {
	Message      chan []byte
	Username     string
	Followers    []string
	NewRoom      chan *TwoUserRoomPayload
	PrivateRooms map[string]*TwoUserRoom
}

func (user *User) ListenForNewRoom() {
	for {
		select {
		case newRoom := <-NewRoom:
			if user.NewRoom != nil {
				user.PrivateRooms[newRoom.Id] = newRoom.Room
			}
		}
	}
}

var OnlineUsers = make(map[string]*User)
var NewUser chan *User
var LoggedOutUser chan *User

func (user *User) ReadFromConnection(connectionKey string) {
	defer socket.Connections[connectionKey].Close()
	connection := socket.Connections[connectionKey]
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
		PrivateRooms[connectionKey].Forward(message)
	}
}

func (user *User) WriteToConnection(connectionKey string) {
	defer socket.Connections[connectionKey].Close()
	connection := socket.Connections[connectionKey]
	for m := range user.Message {
		err := connection.WriteMessage(websocket.TextMessage, m)
		if err != nil {
			fmt.Println("COnnection error: ", err)
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
