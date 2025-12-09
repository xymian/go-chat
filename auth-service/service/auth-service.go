package service

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gorilla/mux"
	pb "github.com/te6lim/go-chat-protos/userpb"
)

var UserService pb.UserServiceClient

var Router *mux.Router = mux.NewRouter()

func ConnectToUserService() error {
	conn, err := grpc.NewClient(
		"user-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	UserService = pb.NewUserServiceClient(conn)

	return nil
}
