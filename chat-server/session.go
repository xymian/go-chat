package chatserver

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type UserSession struct {
	Room      *room
	User      *User
	OtherUser string
}

func CreateSession(user *User, otherUser string) *UserSession {
	var room *room
	if OnlineUsers[otherUser] != nil {
		room = OnlineUsers[otherUser].PrivateRooms[user.Username]
	}

	session := &UserSession{
		Room:      room,
		User:      user,
		OtherUser: otherUser,
	}

	return session
}

func (session *UserSession) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (session *UserSession) ReadMessages() {
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

func (session *UserSession) ForwardMessageToRoom(message Message) {
	session.Room.ForwardedMessage <- message
}

func (session *UserSession) LeaveRoom(user *User) {
	session.Room.leave <- user
}

func (session *UserSession) JoinRoom() error {
	if len(session.Room.participants) < 2 {
		session.Room.join <- session.User
		return nil
	}
	return errors.New("room is full. please create another room with this user")
}
