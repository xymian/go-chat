package routes

import (
	"github.com/te6lim/go-chat/chat-service/controllers"
	"github.com/te6lim/go-chat/chat-service/middleware"
	"github.com/te6lim/go-chat/chat-service/service"
)

func RegisterChatRoutes() {
	service.Router.HandleFunc("/chat", middleware.WithJWTMiddleware(controllers.SetupRoomSocket)).Methods("POST")
	service.Router.HandleFunc("/conversations/{username}", middleware.WithJWTMiddleware(controllers.SetupConversationsSocket)).Methods("POST")

	service.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.InsertMessage)).Methods("POST")
	service.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.DeleteMessage)).Methods("DELETE")
	service.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.GetMessage)).Methods("GET")
	service.Router.HandleFunc("/messages/{chatId}", middleware.WithJWTMiddleware(controllers.DeleteAllMessages)).Methods("DELETE")
	service.Router.HandleFunc("/messages/{chatId}", middleware.WithJWTMiddleware(controllers.GetAllMessages)).Methods("GET")

	service.Router.HandleFunc("/chatReference", middleware.WithJWTMiddleware(controllers.GetChatRefForUsers)).Methods("GET")
	service.Router.HandleFunc("/chatReference", middleware.WithJWTMiddleware(controllers.AddChatReference)).Methods("POST")

	service.Router.HandleFunc("/conversations/{chatReference}", middleware.WithJWTMiddleware(controllers.DeleteConversation)).Methods("DELETE")
	service.Router.HandleFunc("/user/{username}/conversations", middleware.WithJWTMiddleware(controllers.GetUserConversations)).Methods("GET")

	service.Router.HandleFunc("/create-groupchat", middleware.WithJWTMiddleware(controllers.CreateGroupChat)).Methods("POST")
	service.Router.HandleFunc("/group-chat/add-member", middleware.WithJWTMiddleware(controllers.AddGroupMember)).Methods("POST")
	service.Router.HandleFunc("/group-chat/remove-member", middleware.WithJWTMiddleware(controllers.RemoveGroupMember)).Methods("POST")
}
