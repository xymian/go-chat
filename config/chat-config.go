package config

import (
	"fmt"
	"log"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/chat"
	"github.com/te6lim/go-chat/utils"
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
			roomId, err := utils.GenerateUniqueSharedId(user.Username, contact)
			if err != nil {
				log.Fatal(err)
			}
			room := chat.Rooms[roomId]
			if room != nil {
				SetupChat(user.Username, contact, room)
			} else {
				fmt.Println("This user is not on your contact list! Please enter a valid user")
			}
		}
	}
}

func SetupChat(username string, otherUsername string, room *chat.Room) {
	user := chat.OnlineUsers[username]
	if room.ClientConn == nil {
		endpoint := fmt.Sprintf("/%s+%s", username, otherUsername)
		socketURL := fmt.Sprintf("ws://localhost:8080%s", endpoint)
		Router.Handle(endpoint, room)
		conn, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
		if err != nil {
			log.Fatal("WebSocket dial error:", err)
		}
		room.ClientConn = conn
		go room.Run()
	}
	room.JoinRoom(user)
	go room.MessageSender(user)
	go room.MessageReceiver(user)

	for {
		var message string
		fmt.Print("Enter your message: ")
		fmt.Scanln(&message)

		// for testing purposes
		if message == "/" {
			break
		} else {
			user.SendMessage <- chat.SocketMessage{
				Text: message, Sender: username,
			}
		}
	}
}
