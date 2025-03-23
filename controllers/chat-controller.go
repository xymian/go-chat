package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/utils"
)

func InsertMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var message *database.Message
	utils.ParseBody(r, message)
	message = database.InsertMessage(message)
	res, err := json.Marshal(message)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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

/*func DeleteAllMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	otherUserId := r.URL.Query().Get("otherUser")
	chatRef, err := utils.GenerateRoomId(mux.Vars(r)["userId"], otherUserId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	messages := database.DeleteAllMessages(chatRef)
	res, err := json.Marshal(messages)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}*/

func GetMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	messageRef := r.URL.Query().Get("messageId")
	chatRef := r.URL.Query().Get("chatId")
	chat := database.GetChat(chatRef)
	if chat == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	message := database.GetMessage(chat.ChatReference, messageRef)
	res, err := json.Marshal(message)
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
	if chat == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	messages := database.GetAllMessages(chat.ChatReference)
	res, err := json.Marshal(messages)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}
