package templates

import (
	"html/template"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	chat "github.com/te6lim/go-chat/chat"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/utils"
)

type TemplateHandler struct {
	Once     sync.Once
	FileName string
	Template *template.Template
}

type chatPair struct {
	User  string `json:"user"`
	Other string `json:"other"`
}

func (handler *TemplateHandler) parseFileOnce() {
	handler.Once.Do(func() {
		handler.Template = template.Must(template.ParseFiles(filepath.Join("templates", handler.FileName)))
	})
}

func (handler *TemplateHandler) HandleNewChat(w http.ResponseWriter, r *http.Request) {
	handler.parseFileOnce()
	me := r.URL.Query().Get("me")
	user := database.GetUser(me)
	if user == nil {
		_ = database.InsertUser(database.User{
			Username: me,
		})
	}

	data := map[string]interface{}{
		"Host": r.Host,
		"Me":   me,
	}
	handler.Template.Execute(w, data)
}

func (handler *TemplateHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	handler.parseFileOnce()
	roomId := mux.Vars(r)["roomId"]
	chats := database.GetChat(roomId)
	if chats == nil {
		chats = database.InsertChat(database.Chat{
			ChatReference: roomId,
		})
	}

	var pair *chatPair
	utils.ParseBody(r, &pair)

	socketId := uuid.New().String()
	chat.SetupSocketUser(pair.User, pair.Other, chats.ChatReference, socketId)

	data := map[string]interface{}{
		"Host":     r.Host,
		"RoomId":   chats.ChatReference,
		"Pair":     pair,
		"SocketId": socketId,
	}

	handler.Template.Execute(w, data)
}
