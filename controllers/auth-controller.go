package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/utils"
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
	Token string `json:"token"`
}

func RegisterFE(templateHandler *utils.TemplateHandler) http.HandlerFunc {
	templateHandler.ParseFileOnce()
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Host":    r.Host,
			"IsLogin": false,
		}
		templateHandler.Template.Execute(w, data)
	}
}

func LoginFE(templateHandler *utils.TemplateHandler) http.HandlerFunc {
	templateHandler.ParseFileOnce()
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Host":    r.Host,
			"IsLogin": true,
		}
		templateHandler.Template.Execute(w, data)
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	var regRequest registerRequest
	err := utils.ParseBody(r, &regRequest)
	if err != nil {
		response = utils.Error{
			Message: "Parsing body failed",
		}
		w.WriteHeader(http.StatusBadRequest)
	}
	if regRequest.Username == "" || regRequest.Password == "" {
		response = utils.Error{
			Message: "Username and password are required",
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	existingUser := database.GetUser(regRequest.Username)
	if existingUser != nil {
		response = utils.Error{
			Message: "User already exists",
		}
		w.WriteHeader(http.StatusConflict)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	passwordHash, err := utils.HashPassword(regRequest.Password)
	if err != nil {
		response = utils.Error{
			Message: "Hashing password failed",
		}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		user, err := database.InsertUser(
			database.User{
				Username:     regRequest.Username,
				PasswordHash: passwordHash,
			},
		)
		if err != nil {
			response = utils.Error{
				Message: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			response = user
			w.WriteHeader(http.StatusCreated)
		}
	}

	res, _ := json.Marshal(response)
	w.Write(res)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	var response interface{}
	err := utils.ParseBody(r, &request)
	if err != nil {
		response = utils.Error{
			Message: "Parsing body failed",
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	var user = database.GetUser(request.Username)
	if user == nil || !utils.CheckPasswordHash(request.Password, user.PasswordHash) {
		response = utils.Error{
			Message: "Invalid credentials",
		}
		w.WriteHeader(http.StatusUnauthorized)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	token, err := utils.GenerateJWT(user.Username)
	if err != nil {
		response = utils.Error{
			Message: "Generating token failed",
		}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response = loginResponse{
			Token: token,
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
