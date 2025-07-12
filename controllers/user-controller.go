package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/responses"
	"github.com/te6lim/go-chat/utils"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := mux.Vars(r)["username"]
	println("username")
	user := database.GetUser(username)
	var response interface{}
	if user == nil {
		response = responses.Response[string] {
			Data: nil,
			Message: "user does not exist",
			Error: "",
			StatusCode: http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
	} else {
		response = responses.Response[string] {
			Data: &user.Username,
			Message: "",
			Error: "",
			StatusCode: http.StatusOK,
			IsSuccessful: true,
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

func GetInteractions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := mux.Vars(r)["username"]
	user := database.GetUser(username)
	conversationMap := map[int64]string{}
	if user.Interactions != nil {
		err := json.Unmarshal([]byte(*user.Interactions), &conversationMap)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	userIds := []int64{}
	for k := range conversationMap {
		userIds = append(userIds, k)
	}
	users := database.GetUsers(userIds...)
	userChatInfo := []*responses.UserChatInfo{}
	for _, u := range users {
		userChatInfo = append(userChatInfo, &responses.UserChatInfo{
			Username:        u.Username,
			DisplayImageUrl: "",
			ChatReference:   conversationMap[u.Id],
		})
	}

	response := responses.Response[[]*responses.UserChatInfo] {
		Data: &userChatInfo,
		Message: "success",
		Error: "",
		StatusCode: http.StatusOK,
		IsSuccessful: true,
	}

	res, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	users := database.GetAllUsers()
	var response interface{}
	if users == nil {
		response = responses.Response[string] {
			Data: nil,
			Message: "cannot get all users",
			Error: "cannot get all users",
			StatusCode: http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		response = responses.Response[string] {
			Data: nil,
			Message: "",
			Error: "",
			StatusCode: http.StatusOK,
			IsSuccessful: false,
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

func InsertUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	user := &database.User{}
	err := utils.ParseBody(r, &user)
	if err != nil {
		response = responses.Response[string] {
			Data: nil,
			Message: "bad request",
			Error: err.Error(),
			StatusCode: http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	user, err = database.InsertUser(*user)
	if err != nil {
		response = responses.Response[string] {
			Data: nil,
			Message: "cannot insert user",
			Error: err.Error(),
			StatusCode: http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response = responses.Response[string] {
			Data: nil,
			Message: "new user added",
			Error: err.Error(),
			StatusCode: http.StatusOK,
			IsSuccessful: true,
		}
		w.WriteHeader(http.StatusOK)
	}
	res, _ := json.Marshal(response)
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
