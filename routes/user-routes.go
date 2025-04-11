package routes

import (
	"github.com/te6lim/go-chat/config"
	"github.com/te6lim/go-chat/controllers"
	"github.com/te6lim/go-chat/middleware"
)

func RegisterUserRoutes() {
	config.Router.HandleFunc("/user/{username}", middleware.WithJWTMiddleware(controllers.GetUser)).Methods("GET")
	config.Router.HandleFunc("/user/{username}", middleware.WithJWTMiddleware(controllers.Delete)).Methods("DELETE")
	config.Router.HandleFunc("/user", middleware.WithJWTMiddleware(controllers.InsertUser)).Methods("POST")
	config.Router.HandleFunc("/users", middleware.WithJWTMiddleware(controllers.GetAllUsers)).Methods("GET")

	config.Router.HandleFunc("/participant/{username}", middleware.WithJWTMiddleware(controllers.GetParticipant)).Methods("GET")
	config.Router.HandleFunc("/participant", middleware.WithJWTMiddleware(controllers.InsertParticipant)).Methods("POST")
}
