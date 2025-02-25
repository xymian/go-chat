package main

import (
	"fmt"
	"log"
	"net/http"

	chatserver "github.com/te6lim/go-chat/chat-server"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

var Counter = 0

func main() {
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	chatserver.ListenForActiveUsers()

	followers := []string{}
	for f := range chatserver.OnlineUsers {
		followers = append(followers, f)
	}

	newUserId := fmt.Sprint(Counter)

	newUser := &chatserver.User{
		Conn:         conn,
		Message:      make(chan []byte),
		Username:     newUserId,
		Followers:    followers,
		NewRoom:      make(chan *chatserver.TwoUserRoomPayload),
		PrivateRooms: make(map[string]*chatserver.TwoUserRoom),
		RequestToJoinRoom: make(chan string),
	}
	newUser.ListenForJoinRoomRequest()

	newRoom := chatserver.CreateTwoUserRoom()
	defer newRoom.Leave(newUser)

	for _, username := range followers {
		u := chatserver.OnlineUsers[username]
		if u != nil {
			newUser.NewRoom <- &chatserver.TwoUserRoomPayload{
				Room: newRoom, Id: username,
			}
			u.NewRoom <- &chatserver.TwoUserRoomPayload{
				Room: newRoom, Id: newUserId,
			}
			newRoom.Run()
			newRoom.Join(newUser)
			u.RequestToJoinRoom <- newUserId
		}
	}

	go newUser.WriteToConnection()
	newUser.ReadFromConnection()

	chatserver.NewUser <- newUser

	Counter++
}
