package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/models"
	"github.com/te6lim/go-chat/tracer"
)

type Conversations struct {
	PublicConn      *websocket.Conn
	Chats           map[string]bool
	IReceiveMessage chan database.Message
	Tracer          tracer.Tracer
}

func SetUpPublicSocket(username string) {
	endpoint := fmt.Sprintf("/conversations/%s", username)
	config.Router.HandleFunc(endpoint, HandleConversations)
}

func HandleConversations(w http.ResponseWriter, r *http.Request) {
	conn, err := config.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	urlSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	username := urlSegments[1]
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

	socketUser, createErr := CreateSocketUser(user, AWAY)
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

	socketUser.Conversations.IReceiveMessage = make(chan database.Message)
	socketUser.Conversations.PublicConn = conn

	defer func() {
		socketUser.PublicConn.Close()
		socketUser.Tracer.Trace("conversations socket closed")
		LoggedOutUser <- socketUser
	}()

	socketUser.Activity = AWAY
	NewUserFromConversationsSetup <- socketUser
	go socketUser.ReadFromPublicSocket()
	go sendUnacknowledgedMessagesForAllUserCoversations(socketUser)
	socketUser.WriteToPublicSocket()
}

func (user *Socketuser) ReadFromPublicSocket() {
	defer func() {
		user.PublicConn.Close()
		user.Tracer.Trace("conversations socket closed")
		LoggedOutUser <- user
	}()

	message := database.Message{}
	for {
		err := user.PublicConn.ReadJSON(&message)
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
				room.Tracer.Trace(err)
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

func (user *Socketuser) WriteToPublicSocket() {
	defer func() {
		user.PublicConn.Close()
		user.Tracer.Trace("conversations socket closed")
		LoggedOutUser <- user
	}()

	for {
		msg := <-user.IReceiveMessage
		err := user.PublicConn.WriteJSON(msg)
		if err != nil {
			user.Tracer.Trace("Connection error: ", err)
			return
		}
	}
}

func (conversations *Conversations) sendUnacknowledgedMessages(chatRef string, user string) error {
	messages, err := database.GetAllUnacknowledgedMessages(chatRef, user)
	if err != nil {
		return err
	}

	for _, message := range messages {
		conversations.IReceiveMessage <- message
	}
	return nil
}

func sendUnacknowledgedMessagesForAllUserCoversations(socketUser *Socketuser) {
	for chatRef := range socketUser.Chats {
		go func() {
			socketUser.Conversations.sendUnacknowledgedMessages(chatRef, socketUser.Username)
		}()
	}
}
