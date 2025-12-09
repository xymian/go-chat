package service

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat-protos/storage_pb"
	storagepb "github.com/te6lim/go-chat-protos/storage_pb"
	pb "github.com/te6lim/go-chat-protos/userpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	database "github.com/te6lim/go-chat/user-service/database"
)

var StorageService storagepb.StorageServiceClient

var UserServer *Server
var ImageServer *StorageServer

type Server struct {
	pb.UnimplementedUserServiceServer
}
type StorageServer struct {
	storage_pb.UnimplementedStorageServiceServer
}

var Router *mux.Router = mux.NewRouter()

func ConnectToStorageService() error {
	conn, err := grpc.NewClient(
		"storage-service:50054", grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	StorageService = storagepb.NewStorageServiceClient(conn)

	return nil
}

func (server *StorageServer) UpdateUserAvatar(
	grpc.ClientStreamingServer[storage_pb.UpdateUserAvatarRequest, storage_pb.UpdateUserAvatarResponse],
) error {
	return nil
}

func ConnectToUserService() {
	l, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen : %v", err)
	}
	newServer := grpc.NewServer()
	UserServer = &Server{}
	pb.RegisterUserServiceServer(newServer, UserServer)

	fmt.Println("user-service is running on 50052")
	if err := newServer.Serve(l); err != nil {
		log.Fatalf("failed to listen : %v", err)
	}
}



func (server *Server) AddConversation(ctx context.Context, req *pb.AddChatRequest) (*emptypb.Empty, error) {
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

func (server *Server) DeleteUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
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

func (server *Server) InsertUser(ctx context.Context, user *pb.UserResponse) (*pb.UserResponse, error) {
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

func (server *Server) GetUsers(ctx context.Context, empty *emptypb.Empty) (*pb.UserListResponse, error) {
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

func (server *Server) GetUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
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
