package controllers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	chatserver "github.com/te6lim/go-chat/chat-server"
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/database"
)

func HandleTwoUserChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	me := mux.Vars(r)["me"]

	user := database.GetUserdb().GetUser(me)
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

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

	for contact := range user.Contacts {
		fmt.Fprintln(w, "Welcome to chat room with ", contact)

		otherUser := chatserver.OnlineUsers[contact]
		session := chatserver.CreateSession(newUser, contact)
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

	//for testing purposes
	config.ShouldCollectUserInput <- true
}
