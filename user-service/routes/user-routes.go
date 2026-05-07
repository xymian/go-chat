package routes

import (
	"github.com/gorilla/mux"
	controllers "github.com/te6lim/go-chat/user-service/controllers"
	"github.com/te6lim/go-chat/user-service/middleware"
)

var router *mux.Router = mux.NewRouter()

func RegisterUserRoutes() {
	router.HandleFunc("/user/{username}/conversations", middleware.WithJWTMiddleware(controllers.GetConversations)).Methods("GET")
	router.HandleFunc("/user/{username}/profile", middleware.WithJWTMiddleware(controllers.UpdateProfile)).Methods("PATCH")
	router.HandleFunc("/user/{username}", middleware.WithJWTMiddleware(controllers.GetUser)).Methods("GET")
	router.HandleFunc("/user/{username}", middleware.WithJWTMiddleware(controllers.Delete)).Methods("DELETE")
	router.HandleFunc("/user", middleware.WithJWTMiddleware(controllers.InsertUser)).Methods("POST")
	router.HandleFunc("/users", middleware.WithJWTMiddleware(controllers.GetAllUsers)).Methods("GET")
}
