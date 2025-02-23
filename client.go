package main

import (
	"fmt"
	"log"
	"net/http"

	chatserver "github.com/te6lim/go-chat/chat-server"
	"github.com/te6lim/go-chat/socket"
)

func main() {
	chatserver.ListenForActiveUsers()
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := socket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	socket.Connections[fmt.Sprint(socket.Counter)] = conn
	followers := []string{}

	for f := range chatserver.OnlineUsers {
		followers = append(followers, f)
	}

	newUserId := fmt.Sprint(socket.Counter)

	newUser := &chatserver.User{
		Message:      make(chan []byte),
		Username:     newUserId,
		Followers:    followers,
		NewRoom: make(chan *chatserver.TwoUserRoomPayload),
		PrivateRooms: make(map[string]*chatserver.TwoUserRoom),
	}

	for _, username := range followers {
		u := chatserver.OnlineUsers[username]
		if u != nil {
			newRoom := chatserver.CreateTwoUserRoom(u.Username)
			newUser.NewRoom <- &chatserver.TwoUserRoomPayload {
				Room: newRoom, SecondPairId: username,
			}
			u.NewRoom <- &chatserver.TwoUserRoomPayload {
				Room: newRoom, SecondPairId: newUserId,
			}
			newRoom.Join(newUser)
			defer newRoom.Leave()
		}
	}

	go newUser.WriteToConnection(newUserId)
	newUser.ReadFromConnection(newUserId)

	chatserver.NewUser <- newUser

	socket.Counter++
}
