package routes

import (
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/controllers"
	"github.com/te6lim/go-chat/templates"
)

func RegisterChatRoutes(router *mux.Router) {
	router.Handle("/chat", &templates.TemplateHandler{FileName: "start-new-chat.html"}).Methods("GET")
	router.HandleFunc("/chat/{roomId}", controllers.HandleTwoUserChat).Methods("GET")
	router.HandleFunc("/chat/{userId}", controllers.InsertMessage).Methods("POST")
	router.HandleFunc("/chat/{userId}", controllers.DeleteMessage).Methods("DELETE")
	router.HandleFunc("/chat/{userId}", controllers.DeleteAllMessages).Methods("DELETE")
	router.HandleFunc("/chat/{userId}", controllers.GetAllMessages).Methods("GET")
	router.HandleFunc("/chat/{userId}", controllers.GetMessage).Methods("GET")
}
