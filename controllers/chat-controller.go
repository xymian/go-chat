package controllers

import (
	"fmt"
	"net/http"

	//"sync"

	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/chat"
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

	var newUser *chat.User
	if chat.OnlineUsers[me] != nil {
		newUser = chat.OnlineUsers[me]
		newUser.Tracer.Trace("\nUser", me, " is online")
	} else {
		newUser = chat.CreateNewUser(me)
		chat.NewUser <- newUser
		go newUser.ListenForJoinRoomRequest()
	}

	//go newUser.ListenForNewChatSession()

	//var wg *sync.WaitGroup
	for contact := range user.Contacts {
		fmt.Fprintln(w, "Welcome to chat room with ", contact)

		otherUser := chat.OnlineUsers[contact]
		session := chat.CreateSession(newUser, contact)
		newUser.PrivateSessions[session.OtherUser] = session
		newUser.JoinRoomWith(session.OtherUser)
		if otherUser != nil {
			otherUser.RequestToJoinRoom <- newUser.Username
		}
	}

	//for testing purposes
	config.AskForUserToChatWith <- newUser
}
