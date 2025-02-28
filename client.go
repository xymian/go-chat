package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	chatserver "github.com/te6lim/go-chat/chat-server"
	"github.com/te6lim/go-chat/tracer"
)

func main() {
	go chatserver.ListenForActiveUsers()
	http.HandleFunc("/chat", handleTwoUserChat)
	log.Println("Server started on localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleTwoUserChat(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	me := r.URL.Query().Get("me")
	fmt.Fprintln(w, "Welcome to chat room with: ", user)
	
var newUser *chatserver.User
	if chatserver.OnlineUsers[me] != nil {
		newUser = chatserver.OnlineUsers[me]
	} else {
		newUser = &chatserver.User{
			Message:           make(chan chatserver.Message),
			Username:          me,
			Session:           make(chan *chatserver.UserSession),
			PrivateRooms:      make(map[string]*chatserver.TwoUserRoom),
			RequestToJoinRoom: make(chan string),
			Tracer:            tracer.New(),
		}

		chatserver.NewUser <- newUser
	}

	go newUser.ListenForNewRoom()

	otherUser := chatserver.OnlineUsers[user]
	var room *chatserver.TwoUserRoom
	var session *chatserver.UserSession

	if otherUser != nil {
		room = otherUser.PrivateRooms[me]
		session = chatserver.CreateSession(newUser, room, otherUser.Username)
		if room == nil {
			room = chatserver.CreateTwoUserRoom()
			session = chatserver.CreateSession(newUser, room, otherUser.Username)
			endpoint := fmt.Sprintf("/%s+%s", me, user)
			socketURL := fmt.Sprintf("ws://localhost:8080%s", endpoint)
			http.Handle(endpoint, session)
			_, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
			if err != nil {
				log.Fatal("WebSocket dial error:", err)
			}
			go room.Run()
			room.Tracer.Trace("New room added")
		}
		newUser.Session <- session
	} else {
		room = chatserver.CreateTwoUserRoom()
		session = chatserver.CreateSession(newUser, room, user)
		endpoint := fmt.Sprintf("/%s+%s", me, user)
		socketURL := fmt.Sprintf("ws://localhost:8080%s", endpoint)
		http.Handle(endpoint, session)
		_, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
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
			'h', 'e', 'l', 'l', 'o',
		},
		Sender: newUser.Username,
	}
}
