package routes

import (
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/controllers"
)

func RegisterUserRoutes(router *mux.Router) {
	router.HandleFunc("/user/{username}", controllers.GetUser).Methods("GET")
	router.HandleFunc("/user/{username}", controllers.Delete).Methods("DELETE")
	router.HandleFunc("/user", controllers.InsertUser).Methods("POST")
	router.HandleFunc("/users", controllers.GetAllUsers).Methods("GET")

	router.HandleFunc("/participant/{username}", controllers.GetParticipant).Methods("GET")
	router.HandleFunc("/participant", controllers.InsertParticipant).Methods("POST")
}