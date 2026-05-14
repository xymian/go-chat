package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/te6lim/go-chat/chat-service/database"
	"github.com/te6lim/go-chat/chat-service/models"
	"github.com/te6lim/go-chat/chat-service/service"

	pb "github.com/te6lim/go-chat-protos/userpb"
)

type Room struct {
	Id               string
	ChatType         database.ChatType
	leave            chan *Socketuser
	join             chan *Socketuser
	remove           chan string
	participants     map[string]bool
	ForwardedMessage chan database.Message
}

var Rooms map[string]*Room = make(map[string]*Room)
var AddRoom chan *Room = make(chan *Room)

func CreateRoom(roomId string, chatType database.ChatType) *Room {
	room := &Room{
		Id:               roomId,
		ChatType:         chatType,
		leave:            make(chan *Socketuser),
		join:             make(chan *Socketuser),
		remove:           make(chan string),
		participants:     make(map[string]bool),
		ForwardedMessage: make(chan database.Message),
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
			// Mark as AWAY (soft presence) rather than deleting so that
			// messages are still routed to this user's conversations socket
			// while they are on the conversations screen.
			room.participants[user.Username] = false
			LoggedOutFromRoom <- user
			fmt.Println("User", user.Username, " left the room")

		case username := <-room.remove:
			// Hard removal — user was explicitly kicked (group remove, block).
			delete(room.participants, username)
			fmt.Println("User", username, " removed from room")

		case message := <-room.ForwardedMessage:
			// For presence messages in private chats, notify the intended
			// recipient if they are AWAY (conversations socket only). The
			// client always sets ReceiverUsername on presence messages.
			if message.PresenceStatus != nil && room.ChatType == database.ChatTypePrivate && message.ReceiverUsername != nil {
				receiver := GetActiveUser(*message.ReceiverUsername)
				if receiver != nil && receiver.Activity == AWAY && receiver.Notify != nil {
					receiver.Notify <- message
				}
			}

			var toRemove []string
			for username := range room.participants {
				receiver := GetActiveUser(username)
				if receiver == nil {
					// User is fully offline — prune from map lazily.
					toRemove = append(toRemove, username)
					continue
				}
				if receiver.Activity == AWAY {
					if receiver.Notify == nil {
						continue
					}
					// If the conversation was hidden by this user, restore it
					// before delivering the new message so it reappears in
					// their conversation list.
					if !receiver.Chats[message.ChatReference] {
						_, _ = service.UserService.AddUserConversation(
							context.Background(),
							&pb.AddUserConversationRequest{
								UserId:        receiver.UserId,
								ChatReference: message.ChatReference,
								ChatType:      string(room.ChatType),
							},
						)
						receiver.Chats[message.ChatReference] = true
					}
					receiver.Notify <- message
				} else if room.participants[username] {
					receiver.IncomingMessage <- message
				}
			}
			for _, username := range toRemove {
				delete(room.participants, username)
			}
			if len(room.participants) == 0 {
				delete(Rooms, room.Id)
			}
		}
	}
}

func (user *Socketuser) ReadMessages(room *Room, disconnect func()) {
	defer disconnect()
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
			// Block real messages until all participants have accepted the invite.
			hasPending, pendingErr := database.HasPendingParticipant(newMessage.ChatReference)
			if pendingErr == nil && hasPending {
				pendingStatus := "INVITE_PENDING"
				user.PrivateConn.WriteJSON(database.Message{
					MessageReference: newMessage.MessageReference,
					ChatReference:    newMessage.ChatReference,
					SenderUsername:   newMessage.SenderUsername,
					MessageStatus:    &pendingStatus,
					SentTimestamp:    newMessage.SentTimestamp,
				})
				continue
			}
			upToDateMessage, insertErr = database.MaybeInsertAndReturnMostUpToDateMessage(&newMessage)
		}

		if insertErr != nil {
			fmt.Println(err)
			if upToDateMessage != nil {
				room := Rooms[upToDateMessage.ChatReference]
				delete(Rooms, room.Id)
			}
			return
		}

		if upToDateMessage != nil {
			room.ForwardedMessage <- *upToDateMessage
		}
	}
}

func (user *Socketuser) WriteMessages(room *Room, done <-chan struct{}, disconnect func()) {
	defer disconnect()
	for {
		select {
		case message, ok := <-user.IncomingMessage:
			if !ok {
				return
			}
			if room.participants[user.Username] {
				user.PrivateConn.WriteJSON(message)
			} else {
				fmt.Println("You are not in this room")
			}
		case <-done:
			return
		}
	}
}

func (user *Socketuser) LeaveRoom(room *Room) {
	room.leave <- user
}

// Remove sends a hard-removal signal for the given username so that room.Run()
// deletes them from participants immediately. Safe to call from any goroutine.
func (room *Room) Remove(username string) {
	room.remove <- username
}

func (room *Room) JoinRoom(user *Socketuser) error {
	if room.participants[user.Username] {
		return errors.New("user is already in the room")
	}
	if room.ChatType == database.ChatTypePrivate {
		activeCount := 0
		for _, online := range room.participants {
			if online {
				activeCount++
			}
		}
		if activeCount >= 2 {
			return errors.New("room is full. please create another room with this user")
		}
	}
	room.join <- user
	return nil
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

	convsResp, convsErr := service.UserService.GetUserConversations(
		context.Background(),
		&pb.UserRequest{UserId: username},
	)
	if convsErr != nil {
		convsResp = &pb.UserConversationsResponse{}
	}

	newUser, createErr := CreateSocketUser(user, convsResp.Conversations, ONLINE)
	if createErr != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Fatal("error creating socket user")
	}

	// Verify the user is an accepted participant in this chat.
	participant, participantErr := database.GetParticipant(username, chatRef)
	if participantErr != nil || participant == nil || participant.Status != database.ParticipantStatusAccepted {
		w.WriteHeader(http.StatusForbidden)
		response := models.Response[string]{
			Data:         nil,
			Message:      "you are not a participant in this chat",
			Error:        "forbidden",
			StatusCode:   http.StatusForbidden,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	room := Rooms[chatRef]
	if room == nil {
		chat, chatErr := database.GetChat(chatRef)
		if chatErr != nil || chat == nil {
			w.WriteHeader(http.StatusNotFound)
			response := models.Response[string]{
				Data:         nil,
				Message:      "chat not found",
				Error:        "invalid chat reference",
				StatusCode:   http.StatusNotFound,
				IsSuccessful: false,
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
		room = CreateRoom(chatRef, chat.ChatType)
		AddRoom <- room
		go room.Run()
	}

	newUser.PrivateConn = conn

	newUser.Activity = ONLINE
	NewUserFromRoomSetup <- newUser

	// done signals WriteMessages to exit; disconnect runs exactly once so that
	// only one goroutine closes the connection and notifies the room.
	done := make(chan struct{})
	var once sync.Once
	disconnect := func() {
		once.Do(func() {
			close(done)
			newUser.PrivateConn.Close()
			fmt.Println(username, " disconnected")
			room.leave <- newUser
		})
	}

	room.JoinRoom(newUser)
	go newUser.WriteMessages(room, done, disconnect)
	go room.sendUnacknowledgedMessages(chatRef, username)
	newUser.ReadMessages(room, disconnect)
}

func (room *Room) sendUnacknowledgedMessages(chatRef string, user string) error {
	messages, err := database.GetMessagesWithIncompleteReceipts(chatRef, user)
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
