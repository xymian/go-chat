package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/te6lim/go-chat/auth-service/models"
	"github.com/te6lim/go-chat/auth-service/service"
	"github.com/te6lim/go-chat/auth-service/util"

	pb "github.com/te6lim/go-chat-protos/userpb"
)

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiryTime  string `json:"expiryTime"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	var regRequest registerRequest
	err := util.ParseBody(r, &regRequest)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	if regRequest.Username == "" || regRequest.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	existingUser, e := service.UserService.GetUser(
		context.Background(),
		&pb.UserRequest{UserId: regRequest.Username},
	)
	if e != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "error in gRPC call to fetch user",
			Error:        e.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	if existingUser != nil && len(existingUser.Username) != 0 {
		response = models.Response[string]{
			Data:         &existingUser.Username,
			Message:      "User already exists",
			Error:        "",
			StatusCode:   http.StatusConflict,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusConflict)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	passwordHash, err := util.HashPassword(regRequest.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		user, uErr := service.UserService.InsertUser(
			context.Background(),
			&pb.UserResponse{
				Username:     regRequest.Username,
				PasswordHash: passwordHash,
			},
		)
		if uErr != nil {
			response = models.Response[string]{
				Data:         nil,
				Message:      uErr.Error(),
				Error:        uErr.Error(),
				StatusCode:   http.StatusInternalServerError,
				IsSuccessful: false,
			}
			w.WriteHeader(http.StatusInternalServerError)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		} else {
			response = models.Response[string]{
				Data:         &user.Username,
				Message:      "user registered",
				Error:        "",
				StatusCode:   http.StatusOK,
				IsSuccessful: true,
			}
			w.WriteHeader(http.StatusOK)
		}
	}

	res, _ := json.Marshal(response)
	w.Write(res)
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var request loginRequest
	var response interface{}
	err := util.ParseBody(r, &request)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "error in request body",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	var user, e = service.UserService.GetUser(
		context.Background(),
		&pb.UserRequest{UserId: request.Username},
	)
	if e != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "error in gRPC call to fetch user",
			Error:        e.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	if user == nil || !util.CheckPasswordHash(request.Password, user.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		response = models.Response[string]{
			Data:         nil,
			Message:      "user is not authorized",
			Error:        "",
			StatusCode:   http.StatusUnauthorized,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	data, err := util.GenerateJWT(user.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		response = models.Response[loginResponse]{
			Data: &loginResponse{
				AccessToken: data.Token,
				ExpiryTime:  data.Expiry,
			},
			Message:      "login successful",
			Error:        "",
			StatusCode:   http.StatusOK,
			IsSuccessful: true,
		}
		w.WriteHeader(http.StatusOK)
		res, _ := json.Marshal(response)
		w.Write(res)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}
