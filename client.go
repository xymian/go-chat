package main

import (
	"fmt"
	"log"
	"net/http"

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
		Session:           make(chan *chatserver.UserSession),
		PrivateRooms:      make(map[string]*chatserver.TwoUserRoom),
		RequestToJoinRoom: make(chan string),
		Tracer:            tracer.New(),
	}

	chatserver.NewUser <- newUser
	
	go newUser.ListenForNewRoom()

	otherUser := chatserver.OnlineUsers[user]
	var room *chatserver.TwoUserRoom
	var session *chatserver.UserSession

	if otherUser != nil {
		room = otherUser.PrivateRooms[newUserId]
		session = chatserver.CreateSession(newUser, room, otherUser.Username)
		newUser.Session <- session
	} else {
		room = chatserver.CreateTwoUserRoom()
		session = chatserver.CreateSession(newUser, room, user)
		http.Handle("/room", session)
		_, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/room", nil)
		if err != nil {
			log.Fatal("WebSocket dial error:", err)
		}

		newUser.Session <- session
		room.Tracer.Trace("New room added")
		go room.Run()
	}

	session.Join()

	go session.WriteToConnectionWith()
	go session.ReadFromConnectionWith()

	room.ForwardedMessage <- chatserver.Message{
		Text: []byte{
			'h', 'e', 'l', 'l', 'o', byte(Counter),
		},
		Sender: newUser.Username,
	}
	Counter++
}
