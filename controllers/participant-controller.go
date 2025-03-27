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

	var response interface{}
	participant, err = database.InsertParticipant(*participant)
	if err != nil {
		response = utils.Error{
			Err: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
	}
	if participant == nil {
		response = utils.Error{
			Err: "unable to insert participant",
		}
		w.WriteHeader(http.StatusInternalServerError)
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
