package main

import (
	"fmt"
	"log"
	"net/http"

	chatserver "github.com/te6lim/go-chat/chat-server"
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
		newUser.Tracer.Trace("\nUser", me, " is online")
	} else {
		newUser = chatserver.CreateNewUser(me)
		chatserver.NewUser <- newUser
		go newUser.ListenForJoinRoomRequest()
	}

	go newUser.ListenForNewRoom()

	otherUser := chatserver.OnlineUsers[user]
	session := chatserver.CreateSession(newUser, user)
	if otherUser != nil {
		//go func() { otherUser.RequestToJoinRoom <- newUser.Username }()
	}

	go session.ForwardMessageToRoom(
		chatserver.Message{
			Text: []byte{
				'h', 'e', 'l', 'l', 'o',
			},
			Sender: newUser.Username,
		},
	)
}
