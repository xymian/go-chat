package routes

import (
	chatService "github.com/te6lim/go-chat/chat-service"
	"github.com/te6lim/go-chat/chat-service/controllers"
	"github.com/te6lim/go-chat/chat-service/middleware"
	"github.com/te6lim/go-chat/chat-service/util"
)

func RegisterChatRoutes() {
	chatService.Router.HandleFunc(
		"/chat/{username}",
		controllers.HandleNewChat(&util.TemplateHandler{FileName: "start-new-chat.html"}),
	).Methods("GET")
	chatService.Router.HandleFunc(
		"/chat/{username}/{chatId}",
		controllers.HandleChat(&util.TemplateHandler{FileName: "chat.html"}),
	).Methods("GET")

	chatService.Router.HandleFunc("/chat", middleware.WithJWTMiddleware(controllers.SetupRoomSocket)).Methods("POST")
	chatService.Router.HandleFunc("/interactions/{username}", middleware.WithJWTMiddleware(controllers.SetupPublicSocket)).Methods("POST")

	// backup apis. socket takes care of these three
	chatService.Router.HandleFunc("/messages/deliver", middleware.WithJWTMiddleware(controllers.MarkMessagesAsDelivered)).Methods("POST")
	chatService.Router.HandleFunc("/messages/acknowledge", middleware.WithJWTMiddleware(controllers.AcknowledgeMessages)).Methods("POST")
	chatService.Router.HandleFunc("/messages/{chatId}/{username}/unacknowledged", middleware.WithJWTMiddleware(controllers.GetUnacknowledgedMessages)).Methods("GET")
	//

	chatService.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.InsertMessage)).Methods("POST")
	chatService.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.DeleteMessage)).Methods("DELETE")
	chatService.Router.HandleFunc("/messages", middleware.WithJWTMiddleware(controllers.GetMessage)).Methods("GET")
	chatService.Router.HandleFunc("/messages/{chatId}", middleware.WithJWTMiddleware(controllers.DeleteAllMessages)).Methods("DELETE")
	chatService.Router.HandleFunc("/messages/{chatId}", middleware.WithJWTMiddleware(controllers.GetAllMessages)).Methods("GET")

	chatService.Router.HandleFunc("/chatReference", middleware.WithJWTMiddleware(controllers.GetChatRefForUsers)).Methods("GET")
	chatService.Router.HandleFunc("/chatReference", middleware.WithJWTMiddleware(controllers.AddChatReference)).Methods("POST")
}
