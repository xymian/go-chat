package chat

import (
	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/tracer"
)

type room struct {
	Conn             *websocket.Conn
	leave            chan *Socketuser
	join             chan *Socketuser
	participants     map[*Socketuser]bool
	ForwardedMessage chan SocketMessage
	Tracer           tracer.Tracer
}

func CreateTwoUserRoom() *room {
	room := &room{
		leave:            make(chan *Socketuser),
		join:             make(chan *Socketuser),
		participants:     make(map[*Socketuser]bool),
		ForwardedMessage: make(chan SocketMessage),
		Tracer:           tracer.New(),
	}
	return room
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
