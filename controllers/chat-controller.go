package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/utils"
)

func InsertMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var message *database.Message
	var response interface{}
	utils.ParseBody(r, &message)
	message, err := database.InsertMessage(*message)
	if err != nil {
		response = utils.Error{
			Err: err.Error(),
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
			Err: "chat does not exist",
		})
		w.Write(res)
		return
	}
	message := database.GetMessage(chat.ChatReference, messageRef)
	if message == nil {
		response = utils.Error{
			Err: "Message does not exist",
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
	chatRef := r.URL.Query().Get("chatId")
	chat := database.GetChat(chatRef)
	var response interface{}
	if chat == nil {
		res, _ := json.Marshal(utils.Error{
			Err: "chat does not exist",
		})
		w.WriteHeader(http.StatusNotFound)
		w.Write(res)
		return
	}
	messages := database.GetAllMessages(chat.ChatReference)
	if messages == nil {
		response = utils.Error{
			Err: "messages dont exist",
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
			Err: "chat reference does not exists",
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
