package chat

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/tracer"
)

type Room struct {
	Id               string
	ServerConn       *websocket.Conn
	leave            chan *Socketuser
	join             chan *Socketuser
	participants     map[*Socketuser]bool
	ForwardedMessage chan SocketMessage
	Tracer           tracer.Tracer
}

var Rooms map[string]*Room = make(map[string]*Room)
var AddRoom chan *Room = make(chan *Room)

func CreateRoom(roomId string) *Room {
	room := &Room{
		Id:               roomId,
		leave:            make(chan *Socketuser),
		join:             make(chan *Socketuser),
		participants:     make(map[*Socketuser]bool),
		ForwardedMessage: make(chan SocketMessage),
		Tracer:           tracer.New(),
	}
	return room
}

func (room *Room) Run() {
	for {
		select {
		case user := <-room.join:
			room.participants[user] = true
			room.Tracer.Trace("User", user.Username, " joined the room")

		case user := <-room.leave:
			room.participants[user] = false
			delete(room.participants, user)
			close(user.SendMessage)
			if len(room.participants) == 0 {
				delete(Rooms, room.Id)
			}
			room.Tracer.Trace("User", user.Username, " left the room")

		case message := <-room.ForwardedMessage:
			for user := range room.participants {
				user.ReceiveMessage <- message
				room.Tracer.Trace("Forwarded message: ", message.Text, " to User", user.Username)
			}
		}
	}
}

func (room *Room) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	room.ServerConn = conn
	room.ReadMessages()
}

func ListenForNewChatRoom() {
	for room := range AddRoom {
		Rooms[room.Id] = room
		room.Tracer.Trace("new room added to chat")
	}
}

func (room *Room) ReadMessages() {
	defer func() {
		room.ServerConn.Close()
		room.Tracer.Trace("connection closed")
	}()
	for {
		var newMessage *SocketMessage
		err := room.ServerConn.ReadJSON(&newMessage)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
		room.ForwardMessageToRoomMembers(*newMessage)
	}
}

/*func (room *Room) MessageSender(user *Socketuser) {
	if room.participants[user] {
		defer func() {
			user.Tracer.Trace("done sending")
		}()
		for message := range user.SendMessage {
			err := room.ClientConn.WriteJSON(message)
			if err != nil {
				fmt.Println("Connection error: ", err)
				return
			}
		}
	} else {
		user.Tracer.Trace("You are not in this room")
	}
}*/

func (room *Room) MessageReceiver(user *Socketuser) {
	if room.participants[user] {
		defer func() {
			user.Tracer.Trace("done receiving")
		}()
		for message := range user.ReceiveMessage {
			user.Tracer.Trace("message: ", message.Text, "from", message.Sender, "has been received")
			// save message to db or something
		}
	} else {
		user.Tracer.Trace("You are not in this room")
	}
}

func (room *Room) ForwardMessageToRoomMembers(message SocketMessage) {
	room.ForwardedMessage <- message
}

func (room *Room) LeaveRoom(user *Socketuser) {
	room.leave <- user
}

func (room *Room) JoinRoom(user *Socketuser) error {
	if !room.participants[user] {
		if len(room.participants) < 2 {
			room.join <- user
			return nil
		} else {
			return errors.New("room is full. please create another room with this user")
		}
	} else {
		return errors.New("user is already in the room")
	}
}
