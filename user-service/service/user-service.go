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

func (server *Server) AddUserConversation(ctx context.Context, req *pb.AddUserConversationRequest) (*emptypb.Empty, error) {
	_, err := database.AddUserConversation(
		int64(req.UserId),
		req.ChatReference,
		req.ChatType,
		int64(req.OtherUserId),
	)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (server *Server) RemoveUserConversation(ctx context.Context, req *pb.RemoveUserConversationRequest) (*emptypb.Empty, error) {
	err := database.RemoveUserConversation(int64(req.UserId), req.ChatReference)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (server *Server) GetUserConversations(ctx context.Context, req *pb.UserRequest) (*pb.UserConversationsResponse, error) {
	user, err := database.GetUser(req.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &pb.UserConversationsResponse{Conversations: []*pb.UserConversation{}}, nil
	}

	convs, err := database.GetUserConversations(user.Id)
	if err != nil {
		return nil, err
	}

	pbConvs := make([]*pb.UserConversation, 0, len(convs))
	for _, c := range convs {
		pbConvs = append(pbConvs, &pb.UserConversation{
			Id:            uint64(c.Id),
			UserId:        uint64(c.UserId),
			ChatReference: c.ChatReference,
			ChatType:      c.ChatType,
			OtherUserId:   uint64(c.OtherUserId),
			Visible:       c.Visible,
			CreatedAt:     c.CreatedAt,
			UpdatedAt:     c.UpdatedAt,
		})
	}

	return &pb.UserConversationsResponse{Conversations: pbConvs}, nil
}

func (server *Server) DeleteUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	user, err := database.DeleteUser(req.UserId)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return &pb.UserResponse{
			Id:           uint64(user.Id),
			Username:     user.Username,
			PasswordHash: user.PasswordHash,
			DisplayName:  user.DisplayName,
			Bio:          user.Bio,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}, nil
	}

	return nil, nil
}

func (server *Server) InsertUser(ctx context.Context, user *pb.UserResponse) (*pb.UserResponse, error) {
	fmt.Println("has rows: ?")
	insertUser, err := database.InsertUser(
		database.User{
			Username:     user.Username,
			PasswordHash: user.PasswordHash,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		},
	)
	if err != nil {
		return nil, err
	}

	if insertUser != nil {
		return &pb.UserResponse{
			Id:           uint64(insertUser.Id),
			Username:     insertUser.Username,
			PasswordHash: insertUser.PasswordHash,
			DisplayName:  insertUser.DisplayName,
			Bio:          insertUser.Bio,
			CreatedAt:    insertUser.CreatedAt,
			UpdatedAt:    insertUser.UpdatedAt,
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
				Id:           uint64(user.Id),
				Username:     user.Username,
				PasswordHash: user.PasswordHash,
				DisplayName:  user.DisplayName,
				Bio:          user.Bio,
				CreatedAt:    user.CreatedAt,
				UpdatedAt:    user.UpdatedAt,
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
			Id:           uint64(user.Id),
			Username:     user.Username,
			PasswordHash: user.PasswordHash,
			DisplayName:  user.DisplayName,
			Bio:          user.Bio,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}, nil
	}

	return nil, nil
}

func (server *Server) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UserResponse, error) {
	user, err := database.UpdateUserProfile(req.Username, req.DisplayName, req.Bio)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return &pb.UserResponse{
			Id:           uint64(user.Id),
			Username:     user.Username,
			PasswordHash: user.PasswordHash,
			DisplayName:  user.DisplayName,
			Bio:          user.Bio,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
		}, nil
	}
	return nil, nil
}
