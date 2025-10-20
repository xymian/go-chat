package chatservice

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/xymian/go-chat-protos/userpb"
)

var UserService pb.UserServiceClient

var Router *mux.Router = mux.NewRouter()

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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
