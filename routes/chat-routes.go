package routes

import (
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/controllers"
)

func RegisterChatRoutes(router *mux.Router) {
	router.HandleFunc("/chat", controllers.HandleTwoUserChat)
}