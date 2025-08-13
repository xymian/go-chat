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
	endpoint := fmt.Sprintf("/interactions/%s", username)
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

	newSocketUser, errInteractions := SetupSocketUser(user, AWAY)
	if errInteractions != nil {
		w.WriteHeader(http.StatusNotFound)
		response := models.Response[string]{
			Data:         nil,
			Message:      "user does not exist",
			Error:        errInteractions.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	newSocketUser.Conversations.IReceiveMessage = make(chan database.Message)
	newSocketUser.Conversations.PublicConn = conn

	defer func() {
		newSocketUser.PublicConn.Close()
		newSocketUser.Tracer.Trace("conversations socket closed")
		LoggedOutUser <- newSocketUser
	}()

	newSocketUser.Activity = AWAY
	NewUserFromConversationsSetup <- newSocketUser
	go newSocketUser.ReadMessagesFromClient()
	newSocketUser.WriteMessagesToClientSocket()
}

func (conversations *Conversations) ReadMessagesFromClient() {
	defer func() {
		conversations.PublicConn.Close()
	}()

	message := database.Message{}
	for {
		err := conversations.PublicConn.ReadJSON(&message)
		if err != nil {
			return
		}
		room := Rooms[message.ChatReference]
		if room != nil {
			room.ForwardedMessage <- message
		} else {
			_, err := database.InsertMessage(message)
			if err != nil {
				return
			}
		}
	}
}

func (socketUser *Socketuser) WriteMessagesToClientSocket() {
	defer func() {
		socketUser.PublicConn.Close()
		socketUser.Tracer.Trace("conversations socket closed")
	}()

	for {
		msg := <-socketUser.IReceiveMessage
		err := socketUser.PublicConn.WriteJSON(msg)
		if err != nil {
			socketUser.Tracer.Trace("Connection error: ", err)
			return
		}
	}
}
