package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/te6lim/go-chat/models"
	userService "github.com/te6lim/go-chat/user-service"

	"github.com/te6lim/go-chat/utils"
	pb "github.com/xymian/go-chat-protos/userpb"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	username := mux.Vars(r)["username"]
	_, err := userService.UserServer.GetUser(context.Background(), &pb.UserRequest{
		UserId: username,
	})

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
	response = models.Response[pb.UserResponse]{
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
	user, err := userService.UserServer.GetUser(context.Background(), &pb.UserRequest{
		UserId: username,
	})
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
	if len(user.Interactions) != 0 {
		err := json.Unmarshal([]byte(user.Interactions), &conversationMap)
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
	userList, err := userService.UserServer.GetUsers(context.Background(), nil)
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
	for _, u := range userList.Users {
		userChatInfo = append(userChatInfo, &models.UserChatInfo{
			Username:        u.Username,
			DisplayImageUrl: "",
			ChatReference:   conversationMap[int64(u.Id)],
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
	_, err := userService.UserServer.GetUsers(context.Background(), nil)
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
	user := &pb.UserResponse{}
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

	_, uErr := userService.UserServer.InsertUser(context.Background(), user)
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
	user, err := userService.UserServer.DeleteUser(context.Background(), &pb.UserRequest{UserId: username})
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
