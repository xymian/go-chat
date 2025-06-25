package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/chat"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/utils"
)

type addChatReferenceRequest struct {
	User  string
	Other string
}

func HandleChat(templateHandler *utils.TemplateHandler) http.HandlerFunc {
	templateHandler.ParseFileOnce()
	return func(w http.ResponseWriter, r *http.Request) {
		me := mux.Vars(r)["username"]
		chatId := mux.Vars(r)["chatId"]
		var response interface{}
		userChat := database.GetChat(chatId)
		if userChat == nil {
			w.WriteHeader(http.StatusNotFound)
			response = utils.Error{
				Message: "this chat does not exist!",
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
		var participants = database.GetParticipantsInChat(userChat.ChatReference)
		if len(participants) > 2 {
			w.WriteHeader(http.StatusForbidden)
			response = utils.Error{
				Message: "too many participants!",
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
		var other *database.Participant
		var user *database.Participant
		for _, value := range participants {
			if value.Username != me {
				other = value
				break
			}
		}
		if other == nil {
			w.WriteHeader(http.StatusNotFound)
			response = utils.Error{
				Message: "the other participant in this chat dooes not exist!",
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		for _, value := range participants {
			if value.Username == me {
				user = value
				break
			}
		}
		if user == nil {
			w.WriteHeader(http.StatusNotFound)
			response = utils.Error{
				Message: "you are not a participant in this chat!",
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		data := map[string]interface{}{
			"Host":   r.Host,
			"ChatId": userChat.ChatReference,
			"Me":     user.Username,
			"Other":  other.Username,
		}
		chat.SetupSocketUser(me, other.Username, userChat.ChatReference)

		templateHandler.Template.Execute(w, data)
	}
}

func HandleNewChat(templatehandler *utils.TemplateHandler) http.HandlerFunc {
	templatehandler.ParseFileOnce()
	return func(w http.ResponseWriter, r *http.Request) {
		me := mux.Vars(r)["username"]
		data := map[string]interface{}{
			"Host": r.Host,
			"Me":   me,
		}
		templatehandler.Template.Execute(w, data)
	}
}

func InsertMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var message *database.Message
	var response interface{}
	utils.ParseBody(r, &message)
	message, err := database.InsertMessage(*message)
	if err != nil {
		response = utils.Error{
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
	} else {
		response = message
		w.WriteHeader(http.StatusOK)
	}
	res, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	messageReference := r.URL.Query().Get("messageId")
	message := database.DeleteMessage(messageReference)
	res, err := json.Marshal(message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func DeleteAllMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatId"]
	messages := database.DeleteAllMessages(chatRef)
	if messages == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(messages)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func GetMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	messageRef := r.URL.Query().Get("messageId")
	chatRef := r.URL.Query().Get("chatId")
	chat := database.GetChat(chatRef)
	var response interface{}
	if chat == nil {
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(utils.Error{
			Message: "chat does not exist",
		})
		w.Write(res)
		return
	}
	message := database.GetMessage(chat.ChatReference, messageRef)
	if message == nil {
		response = utils.Error{
			Message: "Message does not exist",
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		response = message
		w.WriteHeader(http.StatusOK)
	}
	res, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func GetAllMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatId"]
	chat := database.GetChat(chatRef)
	var response interface{}
	if chat == nil {
		res, _ := json.Marshal(utils.Error{
			Message: "chat does not exist",
		})
		w.WriteHeader(http.StatusNotFound)
		w.Write(res)
		return
	}
	messages := database.GetAllMessages(chat.ChatReference)
	if messages == nil {
		response = utils.Error{
			Message: "messages dont exist",
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		response = messages
		w.WriteHeader(http.StatusOK)
	}
	res, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func GetChatRefForUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := r.URL.Query().Get("user")
	other := r.URL.Query().Get("other")
	chatRef := database.GetChatRefFor(user, other)
	var response interface{}
	if chatRef == nil {
		response = utils.Error{
			Message: "chat reference does not exists",
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		response = map[string]string{
			"chatReference": *chatRef,
		}
		w.WriteHeader(http.StatusOK)
	}
	res, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func AddChatReference(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var refReq addChatReferenceRequest
	var response interface{}
	utils.ParseBody(r, &refReq)
	chatRef := database.GetChatRefFor(refReq.User, refReq.Other)
	if chatRef == nil {
		ref := uuid.NewString()
		_, err := database.InsertParticipant(database.Participant{
			Username:      refReq.User,
			ChatReference: ref,
		})
		if err != nil {
			response := utils.Error{
				Message: err.Error(),
			}
			w.WriteHeader(http.StatusBadRequest)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		_, otherErr := database.InsertParticipant(database.Participant{
			Username:      refReq.Other,
			ChatReference: ref,
		})
		if otherErr != nil {
			response := utils.Error{
				Message: otherErr.Error(),
			}
			w.WriteHeader(http.StatusBadRequest)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		chat, chatErr := database.InsertChat(database.Chat{
			ChatReference: ref,
		})
		if chatErr != nil {
			response := utils.Error{
				Message: chatErr.Error(),
			}
			w.WriteHeader(http.StatusBadRequest)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		response = chat
		w.WriteHeader(http.StatusOK)
		res, _ := json.Marshal(response)
		w.Write(res)
	} else {
		response = chatRef
		w.WriteHeader(http.StatusOK)
		res, _ := json.Marshal(response)
		w.Write(res)
	}
}
