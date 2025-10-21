package routes

import (
	"github.com/te6lim/go-chat/chat-service/service"
	"github.com/te6lim/go-chat/chat-service/controllers"
	"github.com/te6lim/go-chat/chat-service/middleware"
	"github.com/te6lim/go-chat/chat-service/util"
)

func RegisterChatRoutes() {
	service.Router.HandleFunc(
		"/chat/{username}",
		controllers.HandleNewChat(&util.TemplateHandler{FileName: "start-new-chat.html"}),
	).Methods("GET")
	service.Router.HandleFunc(
		"/chat/{username}/{chatId}",
		controllers.HandleChat(&util.TemplateHandler{FileName: "chat.html"}),
	).Methods("GET")

	service.Router.HandleFunc("/chat", middleware.WithJWTMiddleware(controllers.SetupRoomSocket)).Methods("POST")
	service.Router.HandleFunc("/interactions/{username}", middleware.WithJWTMiddleware(controllers.SetupPublicSocket)).Methods("POST")

	// backup apis. socket takes care of these three
	service.Router.HandleFunc("/messages/deliver", middleware.WithJWTMiddleware(controllers.MarkMessagesAsDelivered)).Methods("POST")
	service.Router.HandleFunc("/messages/acknowledge", middleware.WithJWTMiddleware(controllers.AcknowledgeMessages)).Methods("POST")
	service.Router.HandleFunc("/messages/{chatId}/{username}/unacknowledged", middleware.WithJWTMiddleware(controllers.GetUnacknowledgedMessages)).Methods("GET")
	//

	service.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.InsertMessage)).Methods("POST")
	service.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.DeleteMessage)).Methods("DELETE")
	service.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.GetMessage)).Methods("GET")
	service.Router.HandleFunc("/messages/{chatId}", middleware.WithJWTMiddleware(controllers.DeleteAllMessages)).Methods("DELETE")
	service.Router.HandleFunc("/messages/{chatId}", middleware.WithJWTMiddleware(controllers.GetAllMessages)).Methods("GET")

	service.Router.HandleFunc("/chatReference", middleware.WithJWTMiddleware(controllers.GetChatRefForUsers)).Methods("GET")
	service.Router.HandleFunc("/chatReference", middleware.WithJWTMiddleware(controllers.AddChatReference)).Methods("POST")
}
