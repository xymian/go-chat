package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/utils"
)

func InsertParticipant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	participant := &database.Participant{}
	err := utils.ParseBody(r, participant)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	participant = database.InsertParticipant(*participant)
	if participant == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(participant)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func GetParticipant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := r.URL.Query().Get("username")
	chatRef := r.URL.Query().Get("chatReference")

	participant := database.GetParticipant(username, chatRef)
	var response interface{}
	if participant == nil {
		response = utils.Error{
			Err: "participant does not exist",
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		response = participant
		w.WriteHeader(http.StatusOK)
	}
	res, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}
