package config

import (
	"fmt"
	"log"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	chatserver "github.com/te6lim/go-chat/chat-server"
)

var Router *mux.Router = mux.NewRouter()
var Session chan *chatserver.ChatSession = make(chan *chatserver.ChatSession)
var ActiveSessions map[string]*chatserver.ChatSession

var ShouldCollectUserInput = make(chan bool)

func ListenForCollectInputFlag() {
	for shouldCollect := range ShouldCollectUserInput {
		if shouldCollect {
			for {
				var contact string
				fmt.Print("Enter username to chat with: ")
				fmt.Scanln(contact)
				if contact == "/" {
					break
				}
				if ActiveSessions[contact] != nil {
					SetupChat(ActiveSessions[contact])
				} else {
					fmt.Println("This user is not on your contact list! Please enter a valid user")
				}
			}
		}
	}
}

func ListenForActiveSession() {
	for session := range Session {
		ActiveSessions[session.OtherUser] = session
	}
}

func SetupChat(session *chatserver.ChatSession) {
	var sharedConnection = session.User.Connections[session.OtherUser]
	if sharedConnection == nil {
		otherUser := chatserver.OnlineUsers[session.OtherUser]
		if otherUser != nil && otherUser.Connections[session.User.Username] != nil {
			sharedConnection = otherUser.Connections[session.User.Username]
			session.SharedConnection <- sharedConnection
		} else {
			endpoint := fmt.Sprintf("/%s+%s", session.User.Username, session.OtherUser)
			socketURL := fmt.Sprintf("ws://localhost:8080%s", endpoint)
			Router.Handle(endpoint, session)
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
}
