package chat

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type ChatSession struct {
	Room                   *room
	User                   string
	OtherUser              string
	SharedClientConnection *websocket.Conn
}

func CreateSession(user *User, otherUser string) *ChatSession {
	var session *ChatSession
	if OnlineUsers[otherUser] != nil {
		session = OnlineUsers[otherUser].PrivateSessions[user.Username]
	}

	session = &ChatSession{
		Room:      session.Room,
		User:      user.Username,
		OtherUser: otherUser,
	}

	return session
}

func (session *ChatSession) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	user := OnlineUsers[session.User]
	session.Room.Conn = conn
	user.ChatSession <- session
	session.JoinRoom()
	session.ReadMessages()
}

func (session *ChatSession) ReadMessages() {
	defer func() {
		session.Room.Conn.Close()
		session.Room.Tracer.Trace("connection closed")
	}()
	for {
		var newMessage *Message
		err := session.Room.Conn.ReadJSON(&newMessage)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
		session.ForwardMessageToRoom(*newMessage)
	}
}

func (session *ChatSession) ForwardMessageToRoom(message Message) {
	session.Room.ForwardedMessage <- message
}

func (session *ChatSession) LeaveRoom(user *User) {
	session.Room.leave <- user
}

func (session *ChatSession) JoinRoom() error {
	if len(session.Room.participants) < 2 {
		session.Room.join <- OnlineUsers[session.User]
		return nil
	}
	return errors.New("room is full. please create another room with this user")
}
