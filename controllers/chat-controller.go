package controllers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	chatserver "github.com/te6lim/go-chat/chat-server"
	"github.com/te6lim/go-chat/config"
)

func HandleTwoUserChat(w http.ResponseWriter, r *http.Request) {
	other := mux.Vars(r)["user"]
	user := mux.Vars(r)["me"]
	fmt.Fprintln(w, "Welcome to chat room with: ", other)

	var newUser *chatserver.User
	if chatserver.OnlineUsers[user] != nil {
		newUser = chatserver.OnlineUsers[user]
		newUser.Tracer.Trace("\nUser", user, " is online")
	} else {
		newUser = chatserver.CreateNewUser(user)
		chatserver.NewUser <- newUser
		go newUser.ListenForJoinRoomRequest()
	}

	go newUser.ListenForNewRoom()

	otherUser := chatserver.OnlineUsers[other]
	session := chatserver.CreateSession(newUser, other)
	if session.Room == nil {
		session.Room = chatserver.CreateTwoUserRoom()
		go session.Room.Run()
	}
	session.JoinRoom()

	config.Session <- session
	go session.MessageSender()
	go session.MessageReceiver()
	if otherUser != nil {
		otherUser.RequestToJoinRoom <- newUser.Username
	}
}
