package main

import (
	"fmt"
	"log"
	"net/http"

	chatserver "github.com/te6lim/go-chat/chat-server"
)

var sesh chan *chatserver.UserSession = make(chan *chatserver.UserSession)

func listenForActiveSession(action func(s *chatserver.UserSession)) {
	for session := range sesh {
		session.Room.Tracer.Trace("user input set up")
		action(session)
	}
}

func main() {
	go chatserver.ListenForActiveUsers()
	go listenForActiveSession(func(session *chatserver.UserSession) {
		var message string
		fmt.Println("Enter your message: ")
		fmt.Scanln(&message)
		
		session.User.Message <- chatserver.Message{
			Text: message, Sender: session.User.Username,
		}
	})
	http.HandleFunc("/chat", handleTwoUserChat)
	log.Println("Server started on localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("yay")
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
		otherUser.RequestToJoinRoom <- newUser.Username
	}

	session.User.Session <- session
	session.Join()

	go session.ReadMessages()
	go session.WriteMessages()

	sesh <- session
	/* session.ForwardMessageToRoom(
		chatserver.Message{
			Text: []byte{
				'h', 'e', 'l', 'l', 'o',
			},
			Sender: newUser.Username,
		},
	) */

	/* session.Room.Tracer.Trace("sending message...")
	session.User.Message <- chatserver.Message{
		Text:   "hello",
		Sender: newUser.Username,
	} */
}
