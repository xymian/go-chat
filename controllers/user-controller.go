package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/utils"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := mux.Vars(r)["username"]
	user := database.GetUser(username)
	var response interface{}
	if user == nil {
		response = utils.Error{
			Message: "User does not exist",
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		response = user
		w.WriteHeader(http.StatusOK)
	}
	res, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users := database.GetAllUsers()
	var response interface{}
	if users == nil {
		response = utils.Error{
			Message: "User does not exist",
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		response = users
		w.WriteHeader(http.StatusOK)
	}
	res, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func InsertUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := &database.User{}
	err := utils.ParseBody(r, &user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var response interface{}
	user, err = database.InsertUser(*user)
	if err != nil {
		response = utils.Error{
			Message: err.Error(),
		}
		w.WriteHeader(http.StatusBadRequest)
	}
	if user == nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response = user
		w.WriteHeader(http.StatusOK)
	}
	res, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	users := database.DeleteUser(username)
	res, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func DeleteAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users := database.DeleteAllUsers()
	res, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
