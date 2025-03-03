package chatserver

import (
	"fmt"
	"log"
	"net/http"
	"time"

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

	if session.Room == nil {
		session.Room = createTwoUserRoom()
		endpoint := fmt.Sprintf("/%s+%s", user.Username, withUser)
		socketURL := fmt.Sprintf("ws://localhost:8080%s", endpoint)
		http.Handle(endpoint, session)
		_, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
		if err != nil {
			log.Fatal("WebSocket dial error:", err)
		}
		go session.Room.Run()
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
	conn.SetPingHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		return nil
	})
	session.Room.Conn = conn
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

func (session *UserSession) WriteMessages() {
	defer func() {
		session.Room.Conn.Close()
		session.Room.Tracer.Trace("connection closed")
	}()

	for message := range session.User.Message {
		session.Room.Tracer.Trace("writing to room")
		err := session.Room.Conn.WriteJSON(message)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
		session.Room.Tracer.Trace("message was sent...")
	}
}

func (session *UserSession) ForwardMessageToRoom(message Message) {
	session.Room.Forward(message)
}
