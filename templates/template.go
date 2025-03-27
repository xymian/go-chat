package templates

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/chat"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/utils"
)

type TemplateHandler struct {
	Once     sync.Once
	FileName string
	Template *template.Template
}

func (handler *TemplateHandler) parseFileOnce() {
	handler.Once.Do(func() {
		handler.Template = template.Must(template.ParseFiles(filepath.Join("templates", handler.FileName)))
	})
}

func (handler *TemplateHandler) HandleNewChat(w http.ResponseWriter, r *http.Request) {
	handler.parseFileOnce()
	me := r.URL.Query().Get("me")

	var response interface{}
	user := database.GetUser(me)
	if user == nil {
		_, err := database.InsertUser(database.User{
			Username: me,
		})
		if err != nil {
			response = utils.Error{
				Err: "attempted to add an invalid user",
			}
			w.WriteHeader(http.StatusBadRequest)
			res, err := json.Marshal(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(res)
			return
		}
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
	other := r.URL.Query().Get("other")
	var response interface{}
	userChat := database.GetChat(chatId)
	if userChat == nil {
		chat, err := database.InsertChat(database.Chat{
			ChatReference: chatId,
		})
		if err != nil {
			response = utils.Error{
				Err: "attempted to add an invalid chat",
			}
			w.WriteHeader(http.StatusBadRequest)
			res, err := json.Marshal(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(res)
			return
		}
		userChat = chat
	}

	particpant := database.GetParticipant(me, userChat.ChatReference)
	if particpant == nil {
		_, err := database.InsertParticipant(database.Participant{
			Username:      me,
			ChatReference: chatId,
		})
		if err != nil {
			response = utils.Error{
				Err: "attempted to add an invalid participant",
			}
			w.WriteHeader(http.StatusBadRequest)
			res, err := json.Marshal(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(res)
			return
		}
	}

	otherParticipant := database.GetParticipant(other, userChat.ChatReference)
	if otherParticipant == nil {
		_, err := database.InsertParticipant(database.Participant{
			Username:      other,
			ChatReference: chatId,
		})
		if err != nil {
			response = utils.Error{
				Err: "attempted to add an invalid participant",
			}
			w.WriteHeader(http.StatusBadRequest)
			res, err := json.Marshal(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write(res)
			return
		}
	}

	data := map[string]interface{}{
		"Host":   r.Host,
		"ChatId": userChat.ChatReference,
		"Me":     me,
		"Other":  other,
	}
	chat.SetupSocketUser(me, other, userChat.ChatReference)

	handler.Template.Execute(w, data)
}
