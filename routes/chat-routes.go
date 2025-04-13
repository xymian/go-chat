package routes

import (
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/controllers"
	"github.com/te6lim/go-chat/middleware"
	"github.com/te6lim/go-chat/utils"
)

func RegisterChatRoutes() {
	config.Router.HandleFunc(
		"/chat/{username}",
		controllers.HandleNewChat(&utils.TemplateHandler{FileName: "start-new-chat.html"}),
	).Methods("GET")
	config.Router.HandleFunc(
		"/chat/{username}/{chatId}",
		controllers.HandleChat(&utils.TemplateHandler{FileName: "chat.html"}),
	).Methods("GET")
	config.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.InsertMessage)).Methods("POST")
	config.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.DeleteMessage)).Methods("DELETE")
	config.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.GetMessage)).Methods("GET")
	config.Router.HandleFunc("/messages/{chatId}", middleware.WithJWTMiddleware(controllers.DeleteAllMessages)).Methods("DELETE")
	config.Router.HandleFunc("/messages/{chatId}", middleware.WithJWTMiddleware(controllers.GetAllMessages)).Methods("GET")

	config.Router.HandleFunc("/chatReference", middleware.WithJWTMiddleware(controllers.GetChatRefForUsers)).Methods("GET")
	config.Router.HandleFunc("/chatReference", middleware.WithJWTMiddleware(controllers.AddChatReference)).Methods("POST")
}
