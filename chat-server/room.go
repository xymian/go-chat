package chatserver

import (
	"errors"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/tracer"
)

type Room interface {
	Leave(user *User)
	Join(user *User) error
	Forward(channel chan []byte, message []byte)
}

type twoUserRoom struct {
	Conn             *websocket.Conn
	leave            chan *User
	join             chan *User
	participants     map[*User]bool
	ForwardedMessage chan Message
	Tracer           tracer.Tracer
}

func (twoUserRoom *twoUserRoom) Leave(user *User) {
	twoUserRoom.leave <- user
}

func (twoUserRoom *twoUserRoom) Join(user *User) error {
	if len(twoUserRoom.participants) < 2 {
		twoUserRoom.join <- user
		return nil
	}
	return errors.New("Room is full. Please create another room with this user")
}

func (twoUserRoom *twoUserRoom) Forward(message Message) {
	twoUserRoom.ForwardedMessage <- message
}

type MultiUserRoom struct {
}

func (multiUserRoom *MultiUserRoom) Leave(user *User) {
}

func (multiUserRoom *MultiUserRoom) Join(user *User) error {
	return nil
}

func (multiUserRoom *MultiUserRoom) Forward(message []byte) {
}

func createTwoUserRoom() *twoUserRoom {
	return &twoUserRoom{
		leave:            make(chan *User),
		join:             make(chan *User),
		participants:     make(map[*User]bool),
		ForwardedMessage: make(chan Message),
		Tracer:           tracer.New(),
	}
}

func (twoUserRoom *twoUserRoom) Run() {
	for {
		select {
		case user := <-twoUserRoom.join:
			twoUserRoom.participants[user] = true
			twoUserRoom.Tracer.Trace("User", user.Username, " joined the room")

		case user := <-twoUserRoom.leave:
			twoUserRoom.participants[user] = false
			delete(twoUserRoom.participants, user)
			close(user.Message)
			twoUserRoom.Tracer.Trace("User", user.Username, " left the room")

		case message := <-twoUserRoom.ForwardedMessage:
			twoUserRoom.Tracer.Trace("member count: ", len(twoUserRoom.participants))
			for user := range twoUserRoom.participants {
				user.Message <- message
				twoUserRoom.Tracer.Trace("Forwarded message: ", message.Text, " to User", user.Username)
			}
		}
	}
}
