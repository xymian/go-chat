package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/models"
	"github.com/te6lim/go-chat/tracer"
)

type Room struct {
	Id               string
	leave            chan *Socketuser
	join             chan *Socketuser
	participants     map[string]bool
	ForwardedMessage chan database.Message
	Tracer           tracer.Tracer
}

var Rooms map[string]*Room = make(map[string]*Room)
var AddRoom chan *Room = make(chan *Room)

func CreateRoom(roomId string) *Room {
	room := &Room{
		Id:               roomId,
		leave:            make(chan *Socketuser),
		join:             make(chan *Socketuser),
		participants:     make(map[string]bool),
		ForwardedMessage: make(chan database.Message),
		Tracer:           tracer.New(),
	}
	return room
}

func SetupRoomSocket(
	username string, otherUsername string, chatReference string) {

	var room *Room
	if Rooms[chatReference] == nil {
		room = CreateRoom(chatReference)
		AddRoom <- room
		go room.Run()
		endpoint := fmt.Sprintf("/room/%s", chatReference)
		config.Router.HandleFunc(endpoint, room.HandleRoom)
		room.Tracer.Trace("room handler added")
	}
}

func (room *Room) Run() {
	for {
		select {
		case user := <-room.join:
			room.participants[user.Username] = true
			room.Tracer.Trace("User", user.Username, " joined the room")

		case user := <-room.leave:
			room.participants[user.Username] = false
			delete(room.participants, user.Username)
			if len(room.participants) == 0 {
				delete(Rooms, room.Id)
			}
			delete(ActiveSocketUsers, user.Username)
			room.Tracer.Trace("User", user.Username, " left the room")

		case message := <-room.ForwardedMessage:
			_, err := database.InsertMessage(message)
			if err != nil {
				room.Tracer.Trace(err)
				delete(Rooms, room.Id)
				return
			}
			receiver := ActiveSocketUsers[message.ReceiverUsername]
			if receiver != nil {
				if receiver.Activity == AWAY {
					receiver.IReceiveMessage <- message
					room.Tracer.Trace("Message sent through conversation socket: ", message.TextMessage, " to User", receiver.Username)
				}
			}
			for username := range room.participants {
				ActiveSocketUsers[username].ReceiveMessage <- message
				room.Tracer.Trace("Forwarded message: ", message.TextMessage, " to User", username)
			}
		}
	}
}

func (user *Socketuser) ReadMessages(room *Room) {
	defer func() {
		user.PrivateConn.Close()
		user.Tracer.Trace("connection closed")
	}()
	for {
		var newMessage database.Message
		err := user.PrivateConn.ReadJSON(&newMessage)
		if err != nil {
			room.Tracer.Trace("Connection error: ", err)
			return
		}
		room.ForwardedMessage <- newMessage
	}
}

func (user *Socketuser) WriteMessages(room *Room) {
	defer func() {
		user.Tracer.Trace("done receiving")
	}()
	for message := range user.ReceiveMessage {
		if room.participants[user.Username] {
			user.PrivateConn.WriteJSON(message)
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
	if !room.participants[user.Username] {
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

func (room *Room) HandleRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := config.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	username := r.URL.Query().Get("me")
	user, errUser := database.GetUser(username)
	if errUser != nil {
		w.WriteHeader(http.StatusNotFound)
		response := models.Response[string]{
			Data:         nil,
			Message:      "user does not exist",
			Error:        errUser.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	newUser, errInteractions := SetupSocketUser(user, ONLINE)
	if errInteractions != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Fatal("error creating socket user")
	}
	newUser.Activity = ONLINE
	NewUserFromRoomSetup <- newUser

	defer func() {
		newUser.LeaveRoom(room)
		room.Tracer.Trace(username, " disconnected")
		conn.Close()
	}()
	newUser.PrivateConn = conn

	urlSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	chatRef := urlSegments[1]

	room.JoinRoom(newUser)
	go newUser.WriteMessages(room)
	go room.sendUnacknowledgedMessages(chatRef, username)
	newUser.ReadMessages(room)
}

func (room *Room) sendUnacknowledgedMessages(chatRef string, user string) error {
	messages, err := database.GetAllUnacknowledgedMessages(chatRef, user)
	if err != nil {
		return err
	}

	for _, message := range messages {
		room.ForwardedMessage <- message
	}
	return nil
}

func ListenForNewChatRoom() {
	for room := range AddRoom {
		Rooms[room.Id] = room
		room.Tracer.Trace("new room added to chat")
	}
}
