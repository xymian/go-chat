package main

import (
	"fmt"
	"log"
	"net/http"

	chatserver "github.com/te6lim/go-chat/chat-server"

	"github.com/gorilla/websocket"
)

var Counter = 0

func main() {
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	var Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
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
		Message:      make(chan chatserver.Message),
		Username:     newUserId,
		Followers:    followers,
		NewRoom:      make(chan *chatserver.TwoUserRoomPayload),
		PrivateRooms: make(map[string]*chatserver.TwoUserRoom),
		RequestToJoinRoom: make(chan string),
	}
	chatserver.NewUser <- newUser
	newUser.ListenForJoinRoomRequest()

	newRoom := chatserver.CreateTwoUserRoom(conn)
	defer newRoom.Leave(newUser)

	for _, username := range followers {
		otherUser := chatserver.OnlineUsers[username]
		if otherUser != nil {
			newUser.NewRoom <- &chatserver.TwoUserRoomPayload{
				Room: newRoom, Id: username,
			}
			otherUser.NewRoom <- &chatserver.TwoUserRoomPayload{
				Room: newRoom, Id: newUserId,
			}
			newRoom.Run()
			newRoom.Join(newUser)
			otherUser.RequestToJoinRoom <- newUserId
			go newUser.WriteToConnectionWith(otherUser)
			newUser.ReadFromConnectionWith(otherUser)
		}
	}

	Counter++
}
