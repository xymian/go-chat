package chat

import (
	"encoding/json"
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
	participants     map[*Socketuser]bool
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
		participants:     make(map[*Socketuser]bool),
		ForwardedMessage: make(chan database.Message),
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
			delete(OnlineUsers, user.Username)
			room.Tracer.Trace("User", user.Username, " left the room")

		case message := <-room.ForwardedMessage:
			_, err := database.InsertMessage(message)
			if err != nil {
				room.Tracer.Trace(err)
				delete(Rooms, room.Id)
				return
			}
			receiver := OnlineUsers[message.ReceiverUsername]
			if receiver != nil {
				if receiver.Activity == AWAY {
					receiver.ReceiveMessage <- message
				}
			}
			for user := range room.participants {
				user.ReceiveMessage <- message
				room.Tracer.Trace("Forwarded message: ", message.TextMessage, " to User", user.Username)
			}
		}
	}
}

func (room *Room) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	newUser, errInteractions := CreateNewSocketUser(user, ONLINE)
	if errInteractions != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Fatal("error creating socket user")
	}
	NewUser <- newUser

	defer func() {
		newUser.LeaveRoom(room)
		room.Tracer.Trace(username, " disconnected")
		conn.Close()
	}()
	newUser.Conn = conn

	urlSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	chatRef := urlSegments[1]

	room.JoinRoom(newUser)
	go newUser.WriteMessages(room)
	go room.sendUnacknowledgedMessages(chatRef, username)
	newUser.ReadMessages(room)
}

func (room *Room) sendUnacknowledgedMessages(chatRef string, user string) error  {
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
