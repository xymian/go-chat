package chatserver

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type User struct {
	Conn *websocket.Conn
	Message      chan []byte
	Username     string
	Followers    []string
	NewRoom      chan *TwoUserRoomPayload
	PrivateRooms map[string]*TwoUserRoom
	RequestToJoinRoom chan string
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

func (user *User) ReadFromConnection() {
	defer user.Conn.Close()
	for {
		_, message, err := user.Conn.ReadMessage()
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
		PrivateRooms[user.Username].Forward(message)
	}
}

func (user *User) WriteToConnection() {
	defer user.Conn.Close()
	for m := range user.Message {
		err := user.Conn.WriteMessage(websocket.TextMessage, m)
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
	for {
		select {
		case roomWith := <- user.RequestToJoinRoom:
			if user.PrivateRooms[roomWith] != nil {
				user.PrivateRooms[roomWith].Join(user)
			}
		}
	}
}
