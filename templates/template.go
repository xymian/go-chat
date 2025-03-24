package templates

import (
	"html/template"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/chat"
	"github.com/te6lim/go-chat/database"
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
	chatId := mux.Vars(r)["chatId"]
	me := r.URL.Query().Get("me")
	c := database.GetChat(chatId)
	if c == nil {
		c = database.InsertChat(database.Chat{
			ChatReference: chatId,
		})
	}

	particpant := database.GetParticipant(me, c.ChatReference)
	if particpant == nil {
		particpant = database.InsertParticipant(database.Participant{
			Username:      me,
			ChatReference: chatId,
		})
	}

	other := r.URL.Query().Get("other")
	socketId := uuid.New().String()
	data := map[string]interface{}{
		"Host":     r.Host,
		"ChatId":   c.ChatReference,
		"Me":       me,
		"Other":    other,
		"SocketId": socketId,
	}
	chat.SetupSocketUser(me, other, c.ChatReference, socketId)

	handler.Template.Execute(w, data)
}
