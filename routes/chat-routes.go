package routes

import (
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/controllers"
	"github.com/te6lim/go-chat/templates"
)

func RegisterChatRoutes() {
	config.Router.HandleFunc("/chat", (&templates.TemplateHandler{FileName: "start-new-chat.html"}).HandleNewChat).Methods("GET")
	config.Router.HandleFunc("/chat/{chatId}", (&templates.TemplateHandler{FileName: "chat.html"}).HandleChat).Methods("GET")
	config.Router.HandleFunc("/messages", controllers.InsertMessage).Methods("POST")
	config.Router.HandleFunc("/messages", controllers.DeleteMessage).Methods("DELETE")
	config.Router.HandleFunc("/messages", controllers.GetMessage).Methods("GET")
	config.Router.HandleFunc("/messages/{chatId}", controllers.DeleteAllMessages).Methods("DELETE")
	config.Router.HandleFunc("/messages/{chatId}", controllers.GetAllMessages).Methods("GET")

	config.Router.HandleFunc("/chatReference", controllers.GetChatRefForUsers).Methods("GET")
}
