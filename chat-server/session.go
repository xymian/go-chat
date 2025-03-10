package chatserver

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type ChatSession struct {
	Room             *room
	User             *User
	OtherUser        string
	SharedConnection chan *websocket.Conn
}

func CreateSession(user *User, otherUser string) *ChatSession {
	var room *room
	if OnlineUsers[otherUser] != nil {
		room = OnlineUsers[otherUser].PrivateRooms[user.Username]
	}

	session := &ChatSession{
		Room:             room,
		User:             user,
		OtherUser:        otherUser,
		SharedConnection: make(chan *websocket.Conn),
	}

	go session.ListenForSharedConnection()

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

	session.Room.Conn = conn
	session.User.Session <- session
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
		session.Room.join <- session.User
		return nil
	}
	return errors.New("room is full. please create another room with this user")
}

func (session *ChatSession) ListenForSharedConnection() {
	for conn := range session.SharedConnection {
		session.User.Connections[session.OtherUser] = conn
	}
}
