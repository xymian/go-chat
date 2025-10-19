package authservice

import (
	"log"

	controllers "github.com/te6lim/go-chat/auth-service/controllers"
)

func main() {
	var userService, err = ConnectToUserService()
	if err != nil {
		log.Fatal("unable to connect to user-service")
	}
	controllers.UserService = *userService
}
