package routes

import (
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/controllers"
)

func RegisterUserRoutes() {
	config.Router.HandleFunc("/user/{username}", controllers.GetUser).Methods("GET")
	config.Router.HandleFunc("/user/{username}", controllers.Delete).Methods("DELETE")
	config.Router.HandleFunc("/user", controllers.InsertUser).Methods("POST")
	config.Router.HandleFunc("/users", controllers.GetAllUsers).Methods("GET")

	config.Router.HandleFunc("/participant/{username}", controllers.GetParticipant).Methods("GET")
	config.Router.HandleFunc("/participant", controllers.InsertParticipant).Methods("POST")
}