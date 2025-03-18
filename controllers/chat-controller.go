package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/chat"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/utils"
)

func HandleTwoUserChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	me := mux.Vars(r)["me"]

	user := database.GetChatDB().GetUser(me)
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var newUser *chat.Socketuser
	if chat.OnlineUsers[me] != nil {
		newUser = chat.OnlineUsers[me]
		newUser.Tracer.Trace("\nUser", me, " is online")
	} else {
		newUser = chat.CreateNewUser(me)
		chat.NewUser <- newUser
	}

	//for testing purposes
	chat.AskForUserToChatWith <- newUser
}

func InsertMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var message *database.Message
	utils.ParseBody(r, message)
	otherUserId := r.URL.Query().Get("otherUser")
	chatId, err := utils.GenerateRoomId(mux.Vars(r)["userId"], otherUserId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	message = database.GetChatDB().InsertMessage(chatId, message)
	res, err := json.Marshal(message)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Write(res)
}

func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	messageId := r.URL.Query().Get("messageId")
	otherUserId := r.URL.Query().Get("otherUser")
	chatId, err := utils.GenerateRoomId(mux.Vars(r)["userId"], otherUserId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	message := database.GetChatDB().DeleteMessage(chatId, messageId)
	res, err := json.Marshal(message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func DeleteAllMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	otherUserId := r.URL.Query().Get("otherUser")
	chatId, err := utils.GenerateRoomId(mux.Vars(r)["userId"], otherUserId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	messages := database.GetChatDB().DeleteAllMessages(chatId)
	res, err := json.Marshal(messages)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func GetMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	messageId := r.URL.Query().Get("messageId")
	otherUserId := r.URL.Query().Get("otherUser")
	chatId, err := utils.GenerateRoomId(mux.Vars(r)["userId"], otherUserId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	message := database.GetChatDB().GetMessage(chatId, messageId)
	res, err := json.Marshal(message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func GetAllMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	otherUserId := r.URL.Query().Get("otherUser")
	chatId, err := utils.GenerateRoomId(mux.Vars(r)["userId"], otherUserId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	messages := database.GetChatDB().GetAllMessages(chatId)
	res, err := json.Marshal(messages)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}
