package routes

import (
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/controllers"
	"github.com/te6lim/go-chat/middleware"
	"github.com/te6lim/go-chat/templates"
)

func RegisterChatRoutes() {
	config.Router.HandleFunc("/chat", middleware.WithJWTMiddleware((&templates.TemplateHandler{FileName: "start-new-chat.html"}).HandleNewChat)).Methods("GET")
	config.Router.HandleFunc("/chat/{chatId}", middleware.WithJWTMiddleware((&templates.TemplateHandler{FileName: "chat.html"}).HandleChat)).Methods("GET")
	config.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.InsertMessage)).Methods("POST")
	config.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.DeleteMessage)).Methods("DELETE")
	config.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.GetMessage)).Methods("GET")
	config.Router.HandleFunc("/messages/{chatId}", middleware.WithJWTMiddleware(controllers.DeleteAllMessages)).Methods("DELETE")
	config.Router.HandleFunc("/messages/{chatId}", middleware.WithJWTMiddleware(controllers.GetAllMessages)).Methods("GET")

	config.Router.HandleFunc("/chatReference", middleware.WithJWTMiddleware(controllers.GetChatRefForUsers)).Methods("GET")
}
