package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/te6lim/go-chat/user-service/database"
	"github.com/te6lim/go-chat/user-service/models"
	"github.com/te6lim/go-chat/user-service/util"

	service "github.com/te6lim/go-chat/user-service/service"

	pb "github.com/te6lim/go-chat-protos/userpb"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	username := mux.Vars(r)["username"]
	_, err := service.UserServer.GetUser(context.Background(), &pb.UserRequest{
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

func GetConversations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := mux.Vars(r)["username"]
	var response interface{}

	user, err := service.UserServer.GetUser(context.Background(), &pb.UserRequest{
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

	convs, err := database.GetUserConversations(int64(user.Id))
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

	// Build a map of otherUserId -> chatReference for private conversations
	conversationMap := map[int64]string{}
	for _, c := range convs {
		if c.ChatType == "private" {
			conversationMap[c.OtherUserId] = c.ChatReference
		}
	}

	userList, err := service.UserServer.GetUsers(context.Background(), nil)
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
		chatRef := conversationMap[int64(u.Id)]
		if chatRef != "" {
			userChatInfo = append(userChatInfo, &models.UserChatInfo{
				Username:        u.Username,
				DisplayImageUrl: "",
				ChatReference:   chatRef,
			})
		}
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
	_, err := service.UserServer.GetUsers(context.Background(), nil)
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
	err := util.ParseBody(r, &user)
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

	_, uErr := service.UserServer.InsertUser(context.Background(), user)
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

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := mux.Vars(r)["username"]
	var response interface{}

	body := &struct {
		DisplayName string `json:"displayName"`
		Bio         string `json:"bio"`
	}{}
	if err := util.ParseBody(r, body); err != nil {
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

	user, err := service.UserServer.UpdateUserProfile(context.Background(), &pb.UpdateUserProfileRequest{
		Username:    username,
		DisplayName: body.DisplayName,
		Bio:         body.Bio,
	})
	if err != nil || user == nil {
		msg := "user not found"
		if err != nil {
			msg = err.Error()
		}
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not update profile",
			Error:        msg,
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	response = models.Response[pb.UserResponse]{
		Data:         user,
		Message:      "profile updated",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	var response interface{}
	user, err := service.UserServer.DeleteUser(context.Background(), &pb.UserRequest{UserId: username})
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
