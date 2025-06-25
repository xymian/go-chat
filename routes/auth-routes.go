package routes

import (
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/controllers"
	"github.com/te6lim/go-chat/utils"
)

func RegisterAuthRoutes() {
	config.Router.HandleFunc(
		"/register", controllers.RegisterFE(&utils.TemplateHandler{FileName: "auth.html"}),
	).Methods("GET")
	config.Router.HandleFunc(
		"/login", controllers.LoginFE(&utils.TemplateHandler{FileName: "auth.html"}),
	).Methods("GET")

	config.Router.HandleFunc("/register", controllers.Register).Methods("POST")
	config.Router.HandleFunc("/login", controllers.Login).Methods("POST")
	config.Router.HandleFunc("/logout", controllers.Logout).Methods("POST")
}
