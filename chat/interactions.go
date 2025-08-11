package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/models"
	"github.com/te6lim/go-chat/tracer"
)

type Interactions struct {
	MultipleChatConn *websocket.Conn
	Chats            map[string]bool
	IReceiveMessage  chan database.Message
	Tracer           tracer.Tracer
}

func (intrxn *Interactions) Run() {
	for r := range intrxn.IReceiveMessage {
		intrxn.MultipleChatConn.WriteJSON(r)
	}
}

func HandleInteractions(w http.ResponseWriter, r *http.Request) {
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
	}

	newSocketUser, errInteractions := CreateNewSocketUser(user, AWAY)
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
	newSocketUser.Interactions.IReceiveMessage = make(chan database.Message)
	newSocketUser.Interactions.MultipleChatConn = conn

	defer func() {
		conn.Close()
		delete(OnlineUsers, user.Username)
	}()

	newSocketUser.Activity = AWAY
	NewUser <- newSocketUser

	newSocketUser.WriteMessagesToClientSocket()
}

func (socketUser *Socketuser) WriteMessagesToClientSocket() {
	defer func() {
		socketUser.MultipleChatConn.Close()
		socketUser.Tracer.Trace("interactions socket closed")
	}()

	for {
		msg := <- socketUser.IReceiveMessage
		err := socketUser.MultipleChatConn.WriteJSON(msg)
		if err != nil {
			socketUser.Tracer.Trace("Connection error: ", err)
			return
		}
	}
}
