package chatserver

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/tracer"
)

type Room interface {
	Leave(user *User)
	Join(user *User) error
	Forward(channel chan []byte, message []byte)
}

type TwoUserRoom struct {
	Conn             *websocket.Conn
	leave            chan *User
	join             chan *User
	participants     map[*User]bool
	ForwardedMessage chan Message
	Tracer *tracer.EventTracer
}

type TwoUserRoomPayload struct {
	Room *TwoUserRoom
	Id   string
}

func (twoUserRoom *TwoUserRoom) Leave(user *User) {
	twoUserRoom.leave <- user
}

func (twoUserRoom *TwoUserRoom) Join(user *User) error {
	twoUserRoom.join <- user
		/*for member := range twoUserRoom.participants {
			user.WriteToConnectionWith(member)
			user.ReadFromConnectionWith(member)
		}*/
		//twoUserRoom.Leave(user)
	return nil
}

func (twoUserRoom *TwoUserRoom) Forward(message Message) {
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

func CreateTwoUserRoom() *TwoUserRoom {
	return &TwoUserRoom{
		leave:            make(chan *User),
		join:             make(chan *User),
		participants:     make(map[*User]bool),
		ForwardedMessage: make(chan Message),
		Tracer: &tracer.EventTracer{Out: os.Stdout},
	}
}

func (twoUserRoom *TwoUserRoom) Run() {
	for {
		select {
		case user := <-twoUserRoom.join:
			if user != nil {
				twoUserRoom.participants[user] = true
				twoUserRoom.Tracer.Trace("User ", user.Username, " joined the room")
			}

		case user := <-twoUserRoom.leave:
			if user != nil {
				twoUserRoom.participants[user] = false
				delete(twoUserRoom.participants, user)
				close(user.Message)
				twoUserRoom.Tracer.Trace("User ", user.Username, " left the room")
			}

		case message := <-twoUserRoom.ForwardedMessage:
			for user := range twoUserRoom.participants {
				user.Message <- message
				twoUserRoom.Tracer.Trace("Forwarded message: ", message.Text, " to ", user.Username)
			}
		}
	}
}

func (room *TwoUserRoom) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	room.Conn = conn
}
