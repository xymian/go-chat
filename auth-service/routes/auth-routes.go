package routes

import (
	"github.com/xymian/go-chat/auth-service/controllers"
	"github.com/xymian/go-chat/auth-service/util"

	"github.com/gorilla/mux"
)

var router *mux.Router = mux.NewRouter()

func RegisterAuthRoutes() {
	router.HandleFunc(
		"/register", controllers.RegisterFE(&util.TemplateHandler{FileName: "auth.html"}),
	).Methods("GET")
	router.HandleFunc(
		"/login", controllers.LoginFE(&util.TemplateHandler{FileName: "auth.html"}),
	).Methods("GET")

	router.HandleFunc("/register", controllers.Register).Methods("POST")
	router.HandleFunc("/login", controllers.Login).Methods("POST")
	router.HandleFunc("/logout", controllers.Logout).Methods("POST")
}
