package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	chatserver "github.com/te6lim/go-chat/chat-server"
	"github.com/te6lim/go-chat/tracer"
)

var Counter = 0

func main() {
	go chatserver.ListenForActiveUsers()
	http.HandleFunc("/chat", handleWebSocket)
	log.Println("Server started on localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	fmt.Fprintln(w, "Welcome to chat room with: ", user)
	newUserId := fmt.Sprint(Counter)

	newUser := &chatserver.User{
		Message:           make(chan chatserver.Message),
		Username:          newUserId,
		NewRoom:           make(chan *chatserver.TwoUserRoomPayload),
		PrivateRooms:      make(map[string]*chatserver.TwoUserRoom),
		RequestToJoinRoom: make(chan string),
		Tracer:            &tracer.EventTracer{Out: os.Stdout},
	}
	
	go func() {
		chatserver.NewUser <- newUser
	}()
	go newUser.ListenForNewRoom()

	otherUser := chatserver.OnlineUsers[user]
	if otherUser != nil {
		room := otherUser.PrivateRooms[fmt.Sprint(Counter)]
		go func() { 
			newUser.NewRoom <- &chatserver.TwoUserRoomPayload{
				Room: room, Id: user,
			}
		}()
		room.Join(newUser)
	} else {
		room := chatserver.CreateTwoUserRoom()
		go func() { 
			newUser.NewRoom <- &chatserver.TwoUserRoomPayload{
				Room: room, Id: user,
			}
		}()
		http.Handle("/room", room)
		_, _,err := websocket.DefaultDialer.Dial("ws://localhost:8080/room", nil)
		if err != nil {
			log.Fatal("WebSocket dial error:", err)
		}
		go room.Run()
		go room.Join(newUser)
	}
	
	Counter++
}
