package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/responses"
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
		response = responses.Response[string]{
			Data:         nil,
			Message:      "could not add participant",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	if participant == nil {
		response = responses.Response[string]{
			Data:         nil,
			Message:      "unable to insert participant",
			Error:        "unable to insert perticipant",
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	response = responses.Response[database.Participant]{
		Data:         participant,
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)

	res, _ := json.Marshal(response)
	w.Write(res)
}

func GetParticipant(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := r.URL.Query().Get("username")
	chatRef := r.URL.Query().Get("chatReference")

	participant := database.GetParticipant(username, chatRef)
	var response interface{}
	if participant == nil {
		response = responses.Response[string]{
			Data:         nil,
			Message:      "participant does not exist",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response = responses.Response[database.Participant]{
		Data:         participant,
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}
