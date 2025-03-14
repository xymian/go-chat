package config

import (
	"fmt"
	"log"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/chat"
)

var Router *mux.Router = mux.NewRouter()

var AskForUserToChatWith = make(chan *chat.Socketuser)

// for testing purposes
func ListenForCollectInputFlag() {
	for user := range AskForUserToChatWith {
		var contact string
		fmt.Print("Enter username to chat with: ")
		fmt.Scanln(&contact)
		if contact != "/" {
			session := user.PrivateSessions[contact]
			if session != nil {
				SetupChat(session)
			} else {
				fmt.Println("This user is not on your contact list! Please enter a valid user")
			}
		}
	}
}

func SetupChat(session *chat.ChatSession) {
	user := chat.OnlineUsers[session.User]
	var sharedConnection = session.SharedClientConnection
	if sharedConnection == nil {
		otherUser := chat.OnlineUsers[session.OtherUser]
		if otherUser != nil && otherUser.PrivateSessions[session.User] != nil && otherUser.PrivateSessions[session.User].SharedClientConnection != nil {
			sharedConnection = otherUser.PrivateSessions[session.User].SharedClientConnection
			session.SharedClientConnection = sharedConnection
		} else {
			endpoint := fmt.Sprintf("/%s+%s", user.Username, session.OtherUser)
			socketURL := fmt.Sprintf("ws://localhost:8080%s", endpoint)
			Router.Handle(endpoint, session)
			conn, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
			if err != nil {
				log.Fatal("WebSocket dial error:", err)
			}
			session.SharedClientConnection = conn
			go session.Room.Run()
		}
	}
	user.JoinRoomWith(session.OtherUser)
	go session.MessageSender()
	go session.MessageReceiver()

	for {
		var message string
		fmt.Print("Enter your message: ")
		fmt.Scanln(&message)

		// for testing purposes
		if message == "/" {
			break
		} else {
			user.SendMessage <- chat.SocketMessage{
				Text: message, Sender: session.User,
			}
		}
	}
}
