package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/te6lim/go-chat/chat-service/database"
	"github.com/te6lim/go-chat/chat-service/models"
	"github.com/te6lim/go-chat/chat-service/service"

	pb "github.com/te6lim/go-chat-protos/userpb"
)

type Room struct {
	Id               string
	leave            chan *Socketuser
	join             chan *Socketuser
	participants     map[string]bool // users currently connected to this room
	members          map[string]bool // all chat members (for AWAY delivery in groups)
	isGroup          bool
	ForwardedMessage chan database.Message
}

var Rooms map[string]*Room = make(map[string]*Room)
var AddRoom chan *Room = make(chan *Room)

func CreateRoom(roomId string) *Room {
	room := &Room{
		Id:               roomId,
		leave:            make(chan *Socketuser),
		join:             make(chan *Socketuser),
		participants:     make(map[string]bool),
		members:          make(map[string]bool),
		ForwardedMessage: make(chan database.Message),
	}

	// Load group metadata once at room creation.
	chat, err := database.GetChat(roomId)
	if err == nil && chat != nil && chat.IsGroup {
		room.isGroup = true
		participants, pErr := database.GetParticipantsInChat(roomId)
		if pErr == nil {
			for _, p := range participants {
				room.members[p.Username] = true
			}
		}
	}

	return room
}

func SetupRoomSocket(username string, otherUsername string, chatReference string) {
	endpoint := fmt.Sprintf("/room/%s", chatReference)
	service.Router.HandleFunc(endpoint, HandleRoom)
}

func (room *Room) Run() {
	for {
		select {
		case user := <-room.join:
			room.participants[user.Username] = true
			fmt.Println("User", user.Username, " joined the room")

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
				fmt.Println("User", user.Username, " left the room")
				delete(Rooms, room.Id)
			}
			fmt.Println("User", user.Username, " left the room")

		case message := <-room.ForwardedMessage:
			if room.isGroup {
				// Deliver to all group members who are AWAY (not in room right now).
				for username := range room.members {
					if !room.participants[username] {
						awayUser := ActiveSocketUsers[username]
						if awayUser != nil && awayUser.Activity == AWAY {
							awayUser.IReceiveMessage <- message
						}
					}
				}
			} else {
				// DM: deliver to the named receiver if they are AWAY.
				if message.ReceiverUsername != nil {
					receiver := ActiveSocketUsers[*message.ReceiverUsername]
					if receiver != nil && receiver.Activity == AWAY {
						receiver.IReceiveMessage <- message
					}
				}
			}
			// Broadcast to everyone currently in the room.
			for username := range room.participants {
				ActiveSocketUsers[username].ReceiveMessage <- message
			}
		}
	}
}

func (user *Socketuser) ReadMessages(room *Room) {
	defer func() {
		user.PrivateConn.Close()
		fmt.Println("connection closed")
		room.leave <- user
	}()
	for {
		var newMessage database.Message
		err := user.PrivateConn.ReadJSON(&newMessage)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}

		var upToDateMessage *database.Message = &newMessage
		var insertErr error = nil

		switch {
		case newMessage.PresenceStatus == nil && newMessage.MessageStatus == nil:
			upToDateMessage, insertErr = database.MaybeInsertAndReturnMostUpToDateMessage(&newMessage)
		}

		if insertErr != nil {
			fmt.Println(err)
			if upToDateMessage != nil {
				r := Rooms[upToDateMessage.ChatReference]
				delete(Rooms, r.Id)
			}
			return
		}

		if upToDateMessage != nil {
			room.ForwardedMessage <- *upToDateMessage
		}
	}
}

func (user *Socketuser) WriteMessages(room *Room) {
	defer func() {
		fmt.Println("done receiving")
		room.leave <- user
	}()
	for message := range user.ReceiveMessage {
		if room.participants[user.Username] {
			user.PrivateConn.WriteJSON(message)
			// Track group message delivery per-user server-side.
			if room.isGroup {
				_ = database.MarkGroupMessageDelivered(message.MessageReference, user.Username)
			}
		} else {
			fmt.Println("You are not in this room")
		}
	}
}

func (user *Socketuser) LeaveRoom(room *Room) {
	room.leave <- user
}

func (room *Room) JoinRoom(user *Socketuser) error {
	if room.participants[user.Username] {
		return errors.New("user is already in the room")
	}
	if !room.isGroup && len(room.participants) >= 2 {
		return errors.New("room is full. please create another room with this user")
	}
	room.join <- user
	return nil
}

// RefreshMembers reloads group member list from DB (call after add/remove member).
func (room *Room) RefreshMembers() {
	if !room.isGroup {
		return
	}
	participants, err := database.GetParticipantsInChat(room.Id)
	if err != nil {
		return
	}
	newMembers := make(map[string]bool, len(participants))
	for _, p := range participants {
		newMembers[p.Username] = true
	}
	room.members = newMembers
}

func HandleRoom(w http.ResponseWriter, r *http.Request) {
	conn, err := service.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	urlSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	chatRef := urlSegments[1]

	username := r.URL.Query().Get("me")

	user, errUser := service.UserService.GetUser(
		context.Background(),
		&pb.UserRequest{UserId: username},
	)
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
		fmt.Println(username, " disconnected")
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
		fmt.Println("new room added to chat")
	}
}
