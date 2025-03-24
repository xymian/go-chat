package routes

import (
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/controllers"
	"github.com/te6lim/go-chat/templates"
)

func RegisterChatRoutes(router *mux.Router) {
	router.HandleFunc("/chat", (&templates.TemplateHandler{FileName: "start-new-chat.html"}).HandleNewChat).Methods("GET")
	router.HandleFunc("/chat/{chatId}", (&templates.TemplateHandler{FileName: "chat.html"}).HandleChat).Methods("GET")
	router.HandleFunc("/messages", controllers.InsertMessage).Methods("POST")
	router.HandleFunc("/messages", controllers.DeleteMessage).Methods("DELETE")
	router.HandleFunc("/messages", controllers.GetMessage).Methods("GET")
	router.HandleFunc("/messages/{chatId}", controllers.DeleteAllMessages).Methods("DELETE")
	router.HandleFunc("/messages/{chatId}", controllers.GetAllMessages).Methods("GET")
}
