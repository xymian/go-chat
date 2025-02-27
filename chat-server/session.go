package chatserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type UserSession struct {
	room *TwoUserRoom
	user *User
	withUser string
}

func CreateSession(user *User, room *TwoUserRoom, withUser string) *UserSession {
	return &UserSession{
		room: room,
		user: user,
		withUser: withUser,
	}
}

func (session *UserSession) Leave() {
	session.room.Leave(session.user)
}

func (session *UserSession) Join() {
	session.room.Join(session.user)
}

func (session *UserSession) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//defer session.Leave()

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
	session.room.Conn = conn
}

func (session *UserSession) ReadFromConnectionWith() {
	privateRoom := session.user.PrivateRooms[session.withUser]
	defer func() {
		if privateRoom != nil {
			defer privateRoom.Conn.Close()
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
	privateRoom := session.user.PrivateRooms[session.withUser]
	defer func() {
		if privateRoom != nil {
			privateRoom.Conn.Close()
		}
	}()

	for m := range session.user.Message {
		if privateRoom != nil {
			err := privateRoom.Conn.WriteMessage(websocket.TextMessage, m.Text)
			if err != nil {
				fmt.Println("Connection error: ", err)
				return
			}
		}
	}
}