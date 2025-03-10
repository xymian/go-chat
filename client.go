package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	chatserver "github.com/te6lim/go-chat/chat-server"
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/routes"
)

func listenForActiveSession(action func(s *chatserver.UserSession)) {
	for session := range config.Session {
		session.Room.Tracer.Trace("user input set up")
		action(session)
	}
}

func main() {
	var router = mux.NewRouter()
	routes.RegisterUserRoutes(router)
	routes.RegisterChatRoutes(router)

	go chatserver.ListenForActiveUsers()
	go listenForActiveSession(func(session *chatserver.UserSession) {

		var sharedConnection = session.User.Connections[session.OtherUser]
		if sharedConnection == nil {
			otherUser := chatserver.OnlineUsers[session.OtherUser]
			if otherUser != nil && otherUser.Connections[session.User.Username] != nil {
				sharedConnection = otherUser.Connections[session.User.Username]
				session.SharedConnection <- sharedConnection
			} else {
				endpoint := fmt.Sprintf("/%s+%s", session.User.Username, session.OtherUser)
				socketURL := fmt.Sprintf("ws://localhost:8080%s", endpoint)
				router.Handle(endpoint, session)
				conn, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
				if err != nil {
					log.Fatal("WebSocket dial error:", err)
				}
				session.SharedConnection <- conn
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

	log.Println("Server started on localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("server is running")
}
