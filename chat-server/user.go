package chatserver

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type User struct {
	activeMembers map[string]*websocket.Conn
	message       chan []byte
	privateRoom   map[string]*TwoUserRoom
	username      string
}

func (user *User) ReadFromConnection(connectionKey string) {
	defer user.activeMembers[connectionKey].Close()
	connection := user.activeMembers[connectionKey]
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
		user.privateRoom[connectionKey].Forward(message)
	}
}

func (user *User) WriteToConnection(connectionKey string) {
	defer user.activeMembers[connectionKey].Close()
	connection := user.activeMembers[connectionKey]
	for m := range user.message {
		err := connection.WriteMessage(websocket.TextMessage, m)
		if err != nil {
			fmt.Println("COnnection error: ", err)
			return
		}
	}
}
