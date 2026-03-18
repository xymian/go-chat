package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/chat-service/database"
	"github.com/te6lim/go-chat/chat-service/models"
	"github.com/te6lim/go-chat/chat-service/service"

	pb "github.com/te6lim/go-chat-protos/userpb"
)

type Conversations struct {
	Conn   *websocket.Conn
	Chats  map[string]bool
	Notify chan database.Message
}

func SetupConversationsSocket(username string) {
	endpoint := fmt.Sprintf("/conversations/%s", username)
	service.Router.HandleFunc(endpoint, HandleConversations)
}

func HandleConversations(w http.ResponseWriter, r *http.Request) {
	conn, err := service.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	urlSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	username := urlSegments[1]
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

	socketUser, createErr := CreateSocketUser(user, convsResp.Conversations, AWAY)
	if createErr != nil {
		w.WriteHeader(http.StatusNotFound)
		response := models.Response[string]{
			Data:         nil,
			Message:      "user does not exist",
			Error:        createErr.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	socketUser.Conversations.Notify = make(chan database.Message)
	socketUser.Conversations.Conn = conn

	defer func() {
		socketUser.Conn.Close()
		fmt.Println("conversations socket closed")
		LoggedOutUser <- socketUser
	}()

	socketUser.Activity = AWAY
	NewUserFromConversationsSetup <- socketUser
	go socketUser.ReadConversations()
	go sendUnacknowledgedMessagesForAllConversations(socketUser)
	socketUser.WriteConversations()
}

func (user *Socketuser) ReadConversations() {
	defer func() {
		user.Conn.Close()
		fmt.Println("conversations socket closed")
		LoggedOutUser <- user
	}()

	for {
		message := database.Message{}
		err := user.Conn.ReadJSON(&message)
		if err != nil {
			return
		}

		var upToDateMessage *database.Message = &message
		var insertErr error = nil

		switch {
		case message.PresenceStatus == nil && message.MessageStatus == nil:
			upToDateMessage, insertErr = database.MaybeInsertAndReturnMostUpToDateMessage(&message)
		}

		room := Rooms[upToDateMessage.ChatReference]
		if room != nil {
			if insertErr != nil {
				fmt.Println(err)
				delete(Rooms, room.Id)
				return
			}
			room.ForwardedMessage <- *upToDateMessage
		} else {
			if insertErr != nil {
				return
			}
		}
	}
}

func (user *Socketuser) WriteConversations() {
	defer func() {
		user.Conn.Close()
		fmt.Println("conversations socket closed")
		LoggedOutUser <- user
	}()

	for {
		msg := <-user.Notify
		err := user.Conn.WriteJSON(msg)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
	}
}

func (conversations *Conversations) sendUnacknowledgedMessages(chatRef string, user string) error {
	messages, err := database.GetMessagesWithIncompleteReceipts(chatRef, user)
	if err != nil {
		return err
	}

	for _, message := range messages {
		conversations.Notify <- message
	}
	return nil
}

func sendUnacknowledgedMessagesForAllConversations(socketUser *Socketuser) {
	for chatRef := range socketUser.Chats {
		go func() {
			socketUser.Conversations.sendUnacknowledgedMessages(chatRef, socketUser.Username)
		}()
	}
}
