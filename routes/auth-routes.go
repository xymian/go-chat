package routes

import (
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/controllers"
)



func RegisterAuthRoutes() {
	config.Router.HandleFunc("/register", controllers.Register).Methods("POST")
	config.Router.HandleFunc("/login", controllers.Login).Methods("POST")
	config.Router.HandleFunc("/logout", controllers.Logout).Methods("POST")
}