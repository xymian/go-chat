package service

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gorilla/mux"
	pb "github.com/xymian/go-chat-protos/userpb"
)

var Router *mux.Router = mux.NewRouter()

func ConnectToUserService() (*pb.UserServiceClient, error) {
	conn, err := grpc.NewClient(
		"localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewUserServiceClient(conn)

	return &client, nil
}
