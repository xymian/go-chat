package chat

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/tracer"
)

var OnlineUsers = make(map[string]*Socketuser)
var NewUser chan *Socketuser = make(chan *Socketuser)
var LoggedOutUser chan *Socketuser = make(chan *Socketuser)

var AskForUserToChatWith = make(chan *Socketuser)

type PrivateChat struct {
	Conn           *websocket.Conn
	SendMessage    chan database.Message
	ReceiveMessage chan database.Message
}

type Socketuser struct {
	PrivateChat
	*Interactions
	Username string
	Activity Activity
	Tracer   tracer.Tracer
}

func SetupSocketUser(username string, otherUsername string, chatReference string) {
	var room *Room
	if Rooms[chatReference] == nil {
		room = CreateRoom(chatReference)
		AddRoom <- room
		go room.Run()
		endpoint := fmt.Sprintf("/room/%s", chatReference)
		config.Router.Handle(endpoint, room)
		room.Tracer.Trace("room handler added")
	}
}

func SetUpInteractionsSocket(username string) {
	endpoint := fmt.Sprintf("/interactions/%s", username)
	config.Router.HandleFunc(endpoint, HandleInteractions)
}

func CreateNewSocketUser(user *database.User, activity Activity) (*Socketuser, error) {
	interactions := &Interactions{
		Chats: map[string]bool{},
	}
	interactions.Tracer = tracer.New()
	conversationMap := map[int64]string{}
	err := json.Unmarshal([]byte(*user.Interactions), &conversationMap)
	if err != nil {
		return nil, err
	}
	for _, chatRef := range conversationMap {
		interactions.Chats[chatRef] = true
	}
	return &Socketuser{
		Username: user.Username,
		Activity: activity,
		Tracer:   tracer.New(),

		PrivateChat: PrivateChat{
			SendMessage:    make(chan database.Message),
			ReceiveMessage: make(chan database.Message),
		},

		Interactions: interactions,
	}, nil
}

func (user *Socketuser) ReadMessages(room *Room) {
	defer func() {
		user.Conn.Close()
		user.Tracer.Trace("connection closed")
	}()
	for {
		var newMessage *database.Message
		err := user.Conn.ReadJSON(&newMessage)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
		room.ForwardedMessage <- *newMessage
	}
}

func (user *Socketuser) MessageReceiver(room *Room) {
	defer func() {
		user.Tracer.Trace("done receiving")
	}()
	for message := range user.ReceiveMessage {
		if room.participants[user] {
			user.Conn.WriteJSON(message)
			user.Tracer.Trace("message: ", message.TextMessage, "from", message.SenderUsername, "has been received")
		} else {
			user.Tracer.Trace("You are not in this room")
		}
	}
}

func (user *Socketuser) LeaveRoom(room *Room) {
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

func ListenForActiveUsers() {
	for {
		select {
		case newUser := <-NewUser:
			OnlineUsers[newUser.Username] = newUser
			newUser.Tracer.Trace("number of users: ", len(OnlineUsers))
			newUser.Tracer.Trace("New User", newUser.Username, " is online")

		case loggedOutUser := <-LoggedOutUser:
			delete(OnlineUsers, loggedOutUser.Username)
			loggedOutUser.Tracer.Trace("User", loggedOutUser.Username, " logged out")
		}
	}
}
