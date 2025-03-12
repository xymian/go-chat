package chat

import (
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

func CreateSession(user *User, otherUsername string) *ChatSession {
	var room *room
	otherUser := OnlineUsers[otherUsername]
	if otherUser != nil && otherUser.PrivateSessions[user.Username] != nil {
		room = otherUser.PrivateSessions[user.Username].Room
	}

	if room == nil {
		room = CreateTwoUserRoom()
	}

	session := &ChatSession{
		Room:      room,
		User:      user.Username,
		OtherUser: otherUsername,
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
	session.Room.Conn = conn
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
		OnlineUsers[session.User].ForwardMessageToRoom(session.OtherUser, *newMessage)
	}
}
