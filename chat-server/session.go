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

	if (room == nil) {
		room = createTwoUserRoom()
	}

	session := &UserSession{
		Room:     room,
		User:     user,
		WithUser: withUser,
	}

	endpoint := fmt.Sprintf("/%s+%s", user.Username, withUser)
	socketURL := fmt.Sprintf("ws://localhost:8080%s", endpoint)
	http.Handle(endpoint, session)
	_, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
	if err != nil {
		log.Fatal("WebSocket dial error:", err)
	}
	go session.Room.Run()

	session.User.Session <- session
	session.Join()

	go session.WriteToConnectionWith()
	go session.ReadFromConnectionWith()

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
	session.Room.Conn = conn
}

func (session *UserSession) ReadFromConnectionWith() {
	privateRoom := session.User.PrivateRooms[session.WithUser]
	defer func() {
		if privateRoom != nil {
			privateRoom.Conn.Close()
		}
	}()
	for {
		if privateRoom != nil {
			_, message, err := privateRoom.Conn.ReadMessage()
			if err != nil {
				fmt.Println("Connection error: ", err)
				return
			}
			privateRoom.Forward(Message{Text: message, Sender: ""})
		}
	}
}

func (session *UserSession) WriteToConnectionWith() {
	privateRoom := session.User.PrivateRooms[session.WithUser]
	defer func() {
		if privateRoom != nil {
			privateRoom.Conn.Close()
		}
	}()

	for m := range session.User.Message {
		if privateRoom != nil {
			err := privateRoom.Conn.WriteMessage(websocket.TextMessage, m.Text)
			if err != nil {
				fmt.Println("Connection error: ", err)
				return
			}
		}
	}
}

func (session *UserSession) ForwardMessageToRoom(message Message) {
	session.Room.ForwardedMessage <- message
}
