package chatserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type UserSession struct {
	Room     *twoUserRoom
	User     *User
	WithUser string
}

func CreateSession(user *User, withUser string) *UserSession {
	var room *twoUserRoom
	if OnlineUsers[withUser] != nil {
		room = OnlineUsers[withUser].PrivateRooms[user.Username]
	}

	session := &UserSession{
		Room:     room,
		User:     user,
		WithUser: withUser,
	}

	return session
}

func (session *UserSession) Leave() {
	session.Room.Leave(session.User)
}

func (session *UserSession) Join() {
	session.Room.Join(session.User)
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
	session.Join()
	session.ReadMessages()
}

func (session *UserSession) ReadMessages() {
	defer func() {
		session.Room.Conn.Close()
		session.Room.Tracer.Trace("connection closed")
	}()
	for {
		session.Room.Tracer.Trace("read from room")
		var newMessage *Message
		err := session.Room.Conn.ReadJSON(&newMessage)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
		session.Room.Tracer.Trace("message read")
		session.ForwardMessageToRoom(*newMessage)
	}
}

func (session *UserSession) ForwardMessageToRoom(message Message) {
	session.Room.Forward(message)
}
