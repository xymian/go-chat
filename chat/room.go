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

func SetupRoomSocket(username string, otherUsername string, chatReference string) {
	endpoint := fmt.Sprintf("/room/%s", chatReference)
	config.Router.HandleFunc(endpoint, HandleRoom)
}

func (room *Room) Run() {
	for {
		select {
		case user := <-room.join:
			room.participants[user.Username] = true
			room.Tracer.Trace("User", user.Username, " joined the room")

		case user := <-room.leave:
			if ActiveSocketUsers[user.Username] != nil {
				if ActiveSocketUsers[user.Username].PublicConn != nil {
					ActiveSocketUsers[user.Username].Activity = AWAY
				} else {
					LoggedOutUser <- user
				}
			}
			room.participants[user.Username] = false
			delete(room.participants, user.Username)
			if len(room.participants) == 0 {
				room.Tracer.Trace("User", user.Username, " left the room")
				delete(Rooms, room.Id)
			}
			room.Tracer.Trace("User", user.Username, " left the room")

		case message := <-room.ForwardedMessage:
			receiver := ActiveSocketUsers[message.ReceiverUsername]
			//room.Tracer.Trace("receiver on conversations: ", receiver)
			if receiver != nil {
				if receiver.Activity == AWAY {
					receiver.IReceiveMessage <- message
					//room.Tracer.Trace("Message sent through conversation socket: ", message.TextMessage, " to User", receiver.Username)
				}
			}
			for username := range room.participants {
				ActiveSocketUsers[username].ReceiveMessage <- message
				//room.Tracer.Trace("Forwarded message: ", message.TextMessage, " to User", username)
			}
		}
	}
}

func (user *Socketuser) ReadMessages(room *Room) {
	defer func() {
		user.PrivateConn.Close()
		user.Tracer.Trace("connection closed")
		room.leave <- user
	}()
	for {
		var newMessage database.Message
		err := user.PrivateConn.ReadJSON(&newMessage)
		if err != nil {
			user.Tracer.Trace("Connection error: ", err)
			return
		}

		var upToDateMessage *database.Message = &newMessage
		var insertErr error = nil

		switch {
		case newMessage.PresenceStatus == nil && newMessage.MessageStatus == nil:
			upToDateMessage, insertErr = database.MaybeInsertAndReturnMostUpToDateMessage(&newMessage)
			room.Tracer.Trace("message: ", upToDateMessage.TextMessage, " from ", upToDateMessage.SenderUsername, " inserted")
		}

		room := Rooms[upToDateMessage.ChatReference]
		if insertErr != nil {
			room.Tracer.Trace(err)
			delete(Rooms, room.Id)
			return
		}
		room.ForwardedMessage <- *upToDateMessage
	}
}

func (user *Socketuser) WriteMessages(room *Room) {
	defer func() {
		user.Tracer.Trace("done receiving")
		room.leave <- user
	}()
	for message := range user.ReceiveMessage {
		if room.participants[user.Username] {
			user.PrivateConn.WriteJSON(message)
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

func HandleRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := config.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	urlSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	chatRef := urlSegments[1]

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

	newUser, createErr := CreateSocketUser(user, ONLINE)
	if createErr != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Fatal("error creating socket user")
	}

	room := Rooms[chatRef]
	if room == nil {
		room = CreateRoom(chatRef)
		AddRoom <- room
		go room.Run()
	}

	newUser.PrivateConn = conn

	newUser.Activity = ONLINE
	NewUserFromRoomSetup <- newUser

	defer func() {
		room.Tracer.Trace(username, " disconnected")
	}()

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
