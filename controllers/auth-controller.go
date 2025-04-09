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

func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	var regRequest registerRequest
	err := utils.ParseBody(r, &regRequest)
	if err != nil {
		response = utils.Error{
			Err: "Parsing body failed",
		}
		w.WriteHeader(http.StatusBadRequest)
	}
	if regRequest.Username == "" || regRequest.Password == "" {
		response = utils.Error{
			Err: "Username and password are required",
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return

	}
	existingUser := database.GetUser(regRequest.Username)
	if existingUser != nil {
		response = utils.Error{
			Err: "User already exists",
		}
		w.WriteHeader(http.StatusConflict)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	passwordHash, err := utils.HashPassword(regRequest.Password)
	if err != nil {
		response = utils.Error{
			Err: "Hashing password failed",
		}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		user, err := database.InsertUser(
			database.User{
				Username: regRequest.Username,
				Password: passwordHash,
			},
		)
		if err != nil {
			response = utils.Error{
				Err: err.Error(),
			}
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			response = user
			w.WriteHeader(http.StatusCreated)
		}
	}

	res, jsonErr := json.Marshal(response)
	if jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write(res)

}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

}

func Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

}
