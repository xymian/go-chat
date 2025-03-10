package routes

import (
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/controllers"
)

func RegisterUserRoutes(router *mux.Router) {
	router.HandleFunc("/user/{id}", controllers.GetUser).Methods("GET")
	router.HandleFunc("/user/{id}", controllers.Delete).Methods("DELETE")
	router.HandleFunc("/user", controllers.InsertUser).Methods("POST")
	router.HandleFunc("/user", controllers.GetAllUsers).Methods("GET")
}