package service

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/gorilla/mux"
	pb "github.com/xymian/go-chat-protos/userpb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	database "github.com/xymian/go-chat/user-service/database"
)

var UserServer *server

type server struct {
	pb.UnimplementedUserServiceServer
}

var Router *mux.Router = mux.NewRouter()

func ConnectToUserService() {
	l, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen : %v", err)
	}
	newServer := grpc.NewServer()
	UserServer = &server{}
	pb.RegisterUserServiceServer(newServer, UserServer)

	fmt.Println("user-service is running on 50052")
	if err := newServer.Serve(l); err != nil {
		log.Fatalf("failed to listen : %v", err)
	}
}

func (server *server) AddConversation(ctx context.Context, req *pb.AddChatRequest) (*emptypb.Empty, error) {
	otherUser, err := database.GetUser(req.Chat.Username)
	if err != nil {
		return nil, err
	}
	_, err = database.AddConversation(
		&database.User{
			Id:             int64(req.User.Id),
			Username:       req.User.Username,
			PasswordHash:   req.User.PasswordHash,
			ChatReferences: &req.User.ChatReferences,
			CreatedAt:      req.User.CreatedAt,
			UpdatedAt:      req.User.UpdatedAt,
		},
		otherUser.Id,
		req.Chat.ChatRef,
	)

	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (server *server) DeleteUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	user, err := database.DeleteUser(req.UserId)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return &pb.UserResponse{
			Id:             uint64(user.Id),
			Username:       user.Username,
			PasswordHash:   user.PasswordHash,
			ChatReferences: *user.ChatReferences,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
		}, nil
	}

	return nil, nil
}

func (server *server) InsertUser(ctx context.Context, user *pb.UserResponse) (*pb.UserResponse, error) {
	fmt.Println("has rows: ?")
	insertUser, err := database.InsertUser(
		database.User{
			Username:       user.Username,
			PasswordHash:   user.PasswordHash,
			ChatReferences: &user.ChatReferences,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
		},
	)
	if err != nil {
		return nil, err
	}

	if insertUser != nil {
		return &pb.UserResponse{
			Id:             user.Id,
			Username:       user.Username,
			PasswordHash:   user.PasswordHash,
			ChatReferences: user.ChatReferences,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
		}, nil
	}

	return nil, nil
}

func (server *server) GetUsers(ctx context.Context, empty *emptypb.Empty) (*pb.UserListResponse, error) {
	users, err := database.GetAllUsers()
	if err != nil {
		return nil, err
	}
	userList := []*pb.UserResponse{}
	for _, user := range users {
		userList = append(
			userList,
			&pb.UserResponse{
				Id:             uint64(user.Id),
				Username:       user.Username,
				PasswordHash:   user.PasswordHash,
				ChatReferences: *user.ChatReferences,
				CreatedAt:      user.CreatedAt,
				UpdatedAt:      user.UpdatedAt,
			},
		)
	}
	return &pb.UserListResponse{
		Users: userList,
	}, nil
}

func (server *server) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	user, err := database.GetUser(req.UserId)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return &pb.UserResponse{
			Id:             uint64(user.Id),
			Username:       user.Username,
			PasswordHash:   user.PasswordHash,
			ChatReferences: *user.ChatReferences,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
		}, nil
	}

	return nil, nil
}
