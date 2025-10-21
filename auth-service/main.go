package main

import (
	"log"

	controllers "github.com/te6lim/go-chat/auth-service/controllers"
	"github.com/te6lim/go-chat/auth-service/service"
)

func main() {
	var userService, err = service.ConnectToUserService()
	if err != nil {
		log.Fatal("unable to connect to user-service")
	}
	controllers.UserService = *userService
}

/*package main

import (
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/te6lim/go-chat/chat-service/chat"
	"github.com/te6lim/go-chat/chat-service/database"
)

func main() {

	database.ConnectToDB()

	pingErr := database.Instance.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	defer func() {
		database.Instance.Close()
		database.Instance = nil
	}()

	//routes.RegisterUserRoutes()
	//routes.RegisterChatRoutes()
	//routes.RegisterAuthRoutes()

	go chat.ListenForActiveUsers()
	go chat.ListenForNewChatRoom()

	//http.Handle("/", config.Router)
	log.Println("Server started on localhost:6060")
	if err := http.ListenAndServe(":6060", nil); err != nil {
		log.Fatal(err)
	}
}*/
