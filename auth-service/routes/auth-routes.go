package routes

import (
	"github.com/te6lim/go-chat/auth-service/controllers"
	"github.com/te6lim/go-chat/auth-service/service"
)

func RegisterAuthRoutes() {
	service.Router.HandleFunc("/register", controllers.Register).Methods("POST")
	service.Router.HandleFunc("/login", controllers.Login).Methods("POST")
	service.Router.HandleFunc("/logout", controllers.Logout).Methods("POST")
}
