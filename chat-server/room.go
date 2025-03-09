package chatserver

import (
	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/tracer"
)

type room struct {
	Conn             *websocket.Conn
	leave            chan *User
	join             chan *User
	participants     map[*User]bool
	ForwardedMessage chan Message
	Tracer           tracer.Tracer
}

func CreateTwoUserRoom() *room {
	return &room{
		leave:            make(chan *User),
		join:             make(chan *User),
		participants:     make(map[*User]bool),
		ForwardedMessage: make(chan Message),
		Tracer:           tracer.New(),
	}
}

func (room *room) Run() {
	for {
		select {
		case user := <-room.join:
			room.participants[user] = true
			room.Tracer.Trace("User", user.Username, " joined the room")

		case user := <-room.leave:
			room.participants[user] = false
			delete(room.participants, user)
			close(user.SendMessage)
			room.Tracer.Trace("User", user.Username, " left the room")

		case message := <-room.ForwardedMessage:
			for user := range room.participants {
				user.ReceiveMessage <- message
				room.Tracer.Trace("Forwarded message: ", message.Text, " to User", user.Username)
			}
		}
	}
}
