package routes

import (
	"github.com/xymian/go-chat/auth-service/controllers"

	"github.com/gorilla/mux"
)

var router *mux.Router = mux.NewRouter()

func RegisterAuthRoutes() {
	router.HandleFunc("/register", controllers.Register).Methods("POST")
	router.HandleFunc("/login", controllers.Login).Methods("POST")
	router.HandleFunc("/logout", controllers.Logout).Methods("POST")
}