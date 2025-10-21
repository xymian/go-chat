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
		"localhost:6006", grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	return &client, nil
}
