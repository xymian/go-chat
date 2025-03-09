package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
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

		if session.User.Conn == nil {
			otherUser := chatserver.OnlineUsers[session.OtherUser]
			if otherUser != nil && otherUser.Conn != nil {
				session.User.Conn = otherUser.Conn
			} else {
				endpoint := fmt.Sprintf("/%s+%s", session.User.Username, session.OtherUser)
				socketURL := fmt.Sprintf("ws://localhost:8080%s", endpoint)
				http.Handle(endpoint, session)
				conn, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
				if err != nil {
					log.Fatal("WebSocket dial error:", err)
				}
				session.User.Conn = conn
			}
		}

		for {
			var message string
			fmt.Print("Enter your message: ")
			fmt.Scanln(&message)

			if message == "/" {
				break
			} else {
				session.User.SendMessage <- chatserver.Message{
					Text: message, Sender: session.User.Username,
				}
			}
		}
	})
	http.HandleFunc("/chat", handleTwoUserChat)
	log.Println("Server started on localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("server is running")
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
	if session.Room == nil {
		session.Room = chatserver.CreateTwoUserRoom()
		go session.Room.Run()
	}

	sesh <- session
	go session.User.MessageSender()
	go session.User.MessageReceiver()
	if otherUser != nil {
		//otherUser.RequestToJoinRoom <- newUser.Username
	}
}
