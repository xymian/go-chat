package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/models"
	"github.com/te6lim/go-chat/utils"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	username := mux.Vars(r)["username"]
	println("username")
	_, err := database.GetUser(username)

	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "user does not exist",
			Error:        err.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	//TODO: Fix this
	response = models.Response[database.User]{
		Data:         nil,
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)

	res, _ := json.Marshal(response)
	w.Write(res)
}

func GetInteractions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := mux.Vars(r)["username"]
	var response interface{}
	user, err := database.GetUser(username)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "",
			Error:        err.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	conversationMap := map[int64]string{}
	if user.Interactions != nil {
		err := json.Unmarshal([]byte(*user.Interactions), &conversationMap)
		if err != nil {
			response = models.Response[string]{
				Data:         nil,
				Message:      "",
				Error:        err.Error(),
				StatusCode:   http.StatusInternalServerError,
				IsSuccessful: false,
			}
			w.WriteHeader(http.StatusInternalServerError)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
	}

	userIds := []int64{}
	for k := range conversationMap {
		userIds = append(userIds, k)
	}
	users, err := database.GetUsers(userIds...)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "",
			Error:        err.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	userChatInfo := []*models.UserChatInfo{}
	for _, u := range users {
		userChatInfo = append(userChatInfo, &models.UserChatInfo{
			Username:        u.Username,
			DisplayImageUrl: "",
			ChatReference:   conversationMap[u.Id],
		})
	}

	response = models.Response[[]*models.UserChatInfo]{
		Data:         &userChatInfo,
		Message:      "success",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}

	res, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	_, err := database.GetAllUsers()
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "couldn't get all users",
			Error:        err.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	response = models.Response[string]{
		Data:         nil,
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: false,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func InsertUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	user := &database.User{}
	err := utils.ParseBody(r, &user)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "bad request",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	_, uErr := database.InsertUser(*user)
	if uErr != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not insert user",
			Error:        uErr.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response = models.Response[string]{
		Data:         nil,
		Message:      "new user added",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	var response interface{}
	user, err := database.DeleteUser(username)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "",
			Error:        err.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response = models.Response[string]{
		Data:         &user.Username,
		Message:      "user has been deleted",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: false,
	}
	res, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
