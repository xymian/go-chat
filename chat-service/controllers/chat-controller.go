package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/chat-service/chat"
	"github.com/te6lim/go-chat/chat-service/database"
	"github.com/te6lim/go-chat/chat-service/middleware"
	"github.com/te6lim/go-chat/chat-service/models"
	"github.com/te6lim/go-chat/chat-service/service"
	"github.com/te6lim/go-chat/chat-service/util"
	pb "github.com/te6lim/go-chat-protos/userpb"
)

type addChatReferenceRequest struct {
	User  string `json:"user"`
	Other string `json:"other"`
}

type createGroupChatRequest struct {
	Name         string   `json:"name"`
	Participants []string `json:"participants"`
}

type addGroupMemberRequest struct {
	ChatReference string `json:"chatReference"`
	Username      string `json:"username"`
}

func SetupConversationsSocket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := mux.Vars(r)["username"]
	chat.SetupConversationsSocket(username)
	response := models.Response[string]{
		Data:         nil,
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	res, _ := json.Marshal(response)
	w.Write(res)
}

func SetupRoomSocket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newChat = &models.NewChat{}
	util.ParseBody(r, newChat)
	me := newChat.User
	other := newChat.Other
	chatId := newChat.ChatReference
	chat.SetupRoomSocket(me, other, chatId)
	response := models.Response[string]{
		Data:         &chatId,
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	res, _ := json.Marshal(response)
	w.Write(res)
}

func InsertMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var message *database.Message
	var response interface{}
	e := util.ParseBody(r, &message)
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	message, err := database.InsertMessage(*message)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "couldn't add new message",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response = models.Response[database.Message]{
		Data:         message,
		Message:      "message added successfully",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	messageReference := r.URL.Query().Get("messageId")
	message, err := database.DeleteMessage(messageReference)
	if err != nil {
		response = models.Response[database.Message]{
			Data:         nil,
			Message:      "couldn't add new message",
			Error:        err.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	response = models.Response[database.Message]{
		Data:         message,
		Message:      "message deleted successfully",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func DeleteAllMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatId"]
	messages, err := database.DeleteAllMessages(chatRef)
	if err != nil {
		response := models.Response[string]{
			Data:         nil,
			Message:      fmt.Sprintf("%v messages deleted ", len(messages)),
			Error:        err.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: true,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response := models.Response[string]{
		Data:         nil,
		Message:      fmt.Sprintf("%v messages deleted ", len(messages)),
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func GetMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	messageRef := r.URL.Query().Get("messageId")
	chatRef := r.URL.Query().Get("chatId")
	chat, err := database.GetChat(chatRef)
	var response interface{}
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response = models.Response[string]{
			Data:         nil,
			Message:      chatRef + " not found",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	message, err := database.GetMessage(chat.ChatReference, messageRef)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "message not found for " + chatRef,
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response = models.Response[database.Message]{
		Data:         message,
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func GetAllMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatId"]
	chat, err := database.GetChat(chatRef)
	var response interface{}
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      chatRef + " not found",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.WriteHeader(http.StatusNotFound)
		w.Write(res)
		return
	}
	messages, err := database.GetAllMessages(chat.ChatReference)
	if err != nil {
		response = models.Response[[]database.Message]{
			Data:         &messages,
			Message:      "could not get messages",
			Error:        err.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response = models.Response[[]database.Message]{
		Data:         &messages,
		Message:      "messages retrieved",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func GetChatRefForUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := r.URL.Query().Get("user")
	other := r.URL.Query().Get("other")
	chatRef, err := database.GetChatRefFor(user, other)
	var response interface{}
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "chat ref not found",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response = models.Response[string]{
		Data:         chatRef,
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func AddChatReference(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var request addChatReferenceRequest
	var response interface{}
	util.ParseBody(r, &request)

	if request.User == request.Other {
		response := models.Response[string]{
			Data:         nil,
			Message:      "cannot create a chat reference to the same user",
			Error:        "",
			StatusCode:   http.StatusForbidden,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.WriteHeader(http.StatusForbidden)
		w.Write(res)
		return
	}

	otheruser, oErr := service.UserService.GetUser(context.Background(), &pb.UserRequest{UserId: request.Other})
	user, uErr := service.UserService.GetUser(context.Background(), &pb.UserRequest{UserId: request.User})

	if uErr != nil {
		w.WriteHeader(http.StatusNotFound)
		response := models.Response[string]{
			Data:         nil,
			Message:      "",
			Error:        uErr.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	if user == nil || len(user.Username) == 0 {
		w.WriteHeader(http.StatusNotFound)
		response := models.Response[string]{
			Data:         nil,
			Message:      "username " + request.User + " does not exist",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	if oErr != nil {
		w.WriteHeader(http.StatusNotFound)
		response := models.Response[string]{
			Data:         nil,
			Message:      "",
			Error:        oErr.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	if otheruser == nil || len(otheruser.Username) == 0 {
		w.WriteHeader(http.StatusNotFound)
		response := models.Response[string]{
			Data:         nil,
			Message:      "username " + request.Other + " does not exist",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	chatRef, cErr := database.GetChatRefFor(request.User, request.Other)
	if cErr == nil {
		if chatRef == nil {
			ref := uuid.NewString()
			_, pErr := database.InsertParticipant(database.Participant{
				Username:      request.User,
				ChatReference: ref,
			})
			if pErr != nil {
				response = models.Response[string]{
					Data:         nil,
					Message:      "could not add participant to " + ref,
					Error:        pErr.Error(),
					StatusCode:   http.StatusBadRequest,
					IsSuccessful: false,
				}
				w.WriteHeader(http.StatusBadRequest)
				res, _ := json.Marshal(response)
				w.Write(res)
				return
			}

			_, otherErr := database.InsertParticipant(database.Participant{
				Username:      request.Other,
				ChatReference: ref,
			})
			if otherErr != nil {
				response = models.Response[string]{
					Data:         nil,
					Message:      "could not add participant to " + ref,
					Error:        otherErr.Error(),
					StatusCode:   http.StatusBadRequest,
					IsSuccessful: false,
				}
				w.WriteHeader(http.StatusBadRequest)
				res, _ := json.Marshal(response)
				w.Write(res)
				return
			}

			chat, chatErr := database.InsertChat(database.Chat{
				ChatReference: ref,
			})
			if chatErr != nil {
				response = models.Response[string]{
					Data:         nil,
					Message:      "",
					Error:        chatErr.Error(),
					StatusCode:   http.StatusBadRequest,
					IsSuccessful: false,
				}
				w.WriteHeader(http.StatusBadRequest)
				res, _ := json.Marshal(response)
				w.Write(res)
				return
			}

			_, err := service.UserService.AddUserConversation(context.Background(), &pb.AddUserConversationRequest{
				UserId:        uint64(user.Id),
				ChatReference: chat.ChatReference,
				ChatType:      "private",
				OtherUserId:   otheruser.Id,
			})

			if err != nil {
				response = models.Response[string]{
					Data:         nil,
					Message:      "",
					Error:        err.Error(),
					StatusCode:   http.StatusInternalServerError,
					IsSuccessful: false,
				}
				w.WriteHeader(http.StatusInternalServerError)
				res, _ := json.Marshal(response)
				w.Write(res)
				return
			}

			response = models.Response[models.NewChat]{
				Data: &models.NewChat{
					User:          request.User,
					Other:         request.Other,
					ChatReference: chat.ChatReference,
				},
				Message:      "",
				Error:        "",
				StatusCode:   http.StatusOK,
				IsSuccessful: true,
			}
			w.WriteHeader(http.StatusOK)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		} else {
			_, err := service.UserService.AddUserConversation(context.Background(), &pb.AddUserConversationRequest{
				UserId:        uint64(user.Id),
				ChatReference: *chatRef,
				ChatType:      "private",
				OtherUserId:   otheruser.Id,
			})

			if err != nil {
				response = models.Response[string]{
					Data:         nil,
					Message:      "",
					Error:        err.Error(),
					StatusCode:   http.StatusInternalServerError,
					IsSuccessful: false,
				}
				w.WriteHeader(http.StatusInternalServerError)
				res, _ := json.Marshal(response)
				w.Write(res)
				return
			}

			response = models.Response[models.NewChat]{
				Data: &models.NewChat{
					User:          request.User,
					Other:         request.Other,
					ChatReference: *chatRef,
				},
				Message:      "",
				Error:        "",
				StatusCode:   http.StatusOK,
				IsSuccessful: true,
			}

			w.WriteHeader(http.StatusOK)
			res, _ := json.Marshal(response)
			w.Write(res)
		}
	} else {
		println(cErr.Error())
		response = models.Response[string]{
			Data:         nil,
			Message:      "",
			Error:        cErr.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
}

func CreateGroupChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var request createGroupChatRequest
	var response interface{}

	err := util.ParseBody(r, &request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response = models.Response[string]{
			Data:         nil,
			Message:      "invalid request body",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Get the creator from the JWT context
	creator, ok := r.Context().Value(middleware.ContextKeyUsername).(string)
	if !ok || creator == "" {
		w.WriteHeader(http.StatusUnauthorized)
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not identify the requesting user",
			Error:        "unauthorized",
			StatusCode:   http.StatusUnauthorized,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Build full participant list (creator + invited members)
	allParticipants := []string{creator}
	for _, p := range request.Participants {
		if p != creator {
			allParticipants = append(allParticipants, p)
		}
	}

	if len(allParticipants) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		response = models.Response[string]{
			Data:         nil,
			Message:      "a group chat requires at least 3 participants (you + 2 others)",
			Error:        "",
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Validate that all participants exist via gRPC
	validatedUsers := make(map[string]*pb.UserResponse)
	for _, username := range allParticipants {
		user, uErr := service.UserService.GetUser(
			context.Background(), &pb.UserRequest{UserId: username},
		)
		if uErr != nil || user == nil || len(user.Username) == 0 {
			w.WriteHeader(http.StatusNotFound)
			errMsg := ""
			if uErr != nil {
				errMsg = uErr.Error()
			}
			response = models.Response[string]{
				Data:         nil,
				Message:      "user " + username + " does not exist",
				Error:        errMsg,
				StatusCode:   http.StatusNotFound,
				IsSuccessful: false,
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
		validatedUsers[username] = user
	}

	// Generate a new chat reference
	chatRef := uuid.NewString()

	// Insert the chat record with type "group"
	var chatName *string
	if request.Name != "" {
		chatName = &request.Name
	}
	newChat, chatErr := database.InsertChat(database.Chat{
		ChatReference: chatRef,
		ChatType:      database.ChatTypeGroup,
		Name:          chatName,
	})
	if chatErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not create group chat",
			Error:        chatErr.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Insert participants
	for _, username := range allParticipants {
		_, pErr := database.InsertParticipant(database.Participant{
			Username:      username,
			ChatReference: newChat.ChatReference,
		})
		if pErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response = models.Response[string]{
				Data:         nil,
				Message:      "could not add participant " + username,
				Error:        pErr.Error(),
				StatusCode:   http.StatusInternalServerError,
				IsSuccessful: false,
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
	}

	// Register the group chat reference on each participant via gRPC
	for _, username := range allParticipants {
		user := validatedUsers[username]
		_, grpcErr := service.UserService.AddUserConversation(
			context.Background(),
			&pb.AddUserConversationRequest{
				UserId:        user.Id,
				ChatReference: newChat.ChatReference,
				ChatType:      "group",
			},
		)
		if grpcErr != nil {
			fmt.Println("failed to register group chat ref for", username, ":", grpcErr)
		}
	}

	// Notify invited users (everyone except the creator) via their public socket
	groupInviteStatus := "GROUP_INVITE"
	for _, username := range allParticipants {
		if username == creator {
			continue
		}
		activeUser := chat.ActiveSocketUsers[username]
		if activeUser != nil && activeUser.Notify != nil {
			activeUser.Notify <- database.Message{
				MessageReference: uuid.NewString(),
				SenderUsername:   creator,
				ChatReference:    newChat.ChatReference,
				MessageStatus:    &groupInviteStatus,
				SentTimestamp:    time.Now().Format(time.RFC3339),
			}
		}
	}

	// Setup the room socket for this group chat
	chat.SetupRoomSocket(creator, "", newChat.ChatReference)

	// Return success
	response = models.Response[models.NewGroupChat]{
		Data: &models.NewGroupChat{
			ChatReference: newChat.ChatReference,
			Name:          newChat.Name,
			Participants:  allParticipants,
		},
		Message:      "group chat created successfully",
		Error:        "",
		StatusCode:   http.StatusCreated,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusCreated)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func AddGroupMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var request addGroupMemberRequest
	var response interface{}

	err := util.ParseBody(r, &request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response = models.Response[string]{
			Data:         nil,
			Message:      "invalid request body",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Get the requester from the JWT context
	requester, ok := r.Context().Value(middleware.ContextKeyUsername).(string)
	if !ok || requester == "" {
		w.WriteHeader(http.StatusUnauthorized)
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not identify the requesting user",
			Error:        "unauthorized",
			StatusCode:   http.StatusUnauthorized,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Validate chat exists and is a group chat
	groupChat, chatErr := database.GetChat(request.ChatReference)
	if chatErr != nil || groupChat == nil {
		w.WriteHeader(http.StatusNotFound)
		response = models.Response[string]{
			Data:         nil,
			Message:      "chat not found",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	if groupChat.ChatType != database.ChatTypeGroup {
		w.WriteHeader(http.StatusBadRequest)
		response = models.Response[string]{
			Data:         nil,
			Message:      "members can only be added to group chats",
			Error:        "",
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Validate requester is a member of this group
	requesterParticipant, _ := database.GetParticipant(requester, request.ChatReference)
	if requesterParticipant == nil {
		w.WriteHeader(http.StatusForbidden)
		response = models.Response[string]{
			Data:         nil,
			Message:      "you are not a member of this group chat",
			Error:        "",
			StatusCode:   http.StatusForbidden,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Validate the new user exists via gRPC
	newUser, uErr := service.UserService.GetUser(
		context.Background(), &pb.UserRequest{UserId: request.Username},
	)
	if uErr != nil || newUser == nil || len(newUser.Username) == 0 {
		w.WriteHeader(http.StatusNotFound)
		errMsg := ""
		if uErr != nil {
			errMsg = uErr.Error()
		}
		response = models.Response[string]{
			Data:         nil,
			Message:      "user " + request.Username + " does not exist",
			Error:        errMsg,
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Check if the user is already a participant
	existingParticipant, _ := database.GetParticipant(request.Username, request.ChatReference)
	if existingParticipant != nil {
		w.WriteHeader(http.StatusConflict)
		response = models.Response[string]{
			Data:         nil,
			Message:      request.Username + " is already a member of this group",
			Error:        "",
			StatusCode:   http.StatusConflict,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Insert participant
	_, pErr := database.InsertParticipant(database.Participant{
		Username:      request.Username,
		ChatReference: request.ChatReference,
	})
	if pErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not add participant",
			Error:        pErr.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Register the group chat reference on the user via gRPC
	_, grpcErr := service.UserService.AddUserConversation(
		context.Background(),
		&pb.AddUserConversationRequest{
			UserId:        newUser.Id,
			ChatReference: request.ChatReference,
			ChatType:      "group",
		},
	)
	if grpcErr != nil {
		fmt.Println("failed to register group chat ref for", request.Username, ":", grpcErr)
	}

	// Notify the new member via their public socket
	groupInviteStatus := "GROUP_INVITE"
	activeUser := chat.ActiveSocketUsers[request.Username]
	if activeUser != nil && activeUser.Notify != nil {
		activeUser.Notify <- database.Message{
			MessageReference: uuid.NewString(),
			SenderUsername:   requester,
			ChatReference:    request.ChatReference,
			MessageStatus:    &groupInviteStatus,
			SentTimestamp:    time.Now().Format(time.RFC3339),
		}
	}

	// Fetch updated participant list for the response
	participants, _ := database.GetParticipantsInChat(request.ChatReference)
	participantNames := []string{}
	for _, p := range participants {
		participantNames = append(participantNames, p.Username)
	}

	response = models.Response[models.NewGroupChat]{
		Data: &models.NewGroupChat{
			ChatReference: request.ChatReference,
			Name:          groupChat.Name,
			Participants:  participantNames,
		},
		Message:      request.Username + " added to group",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var request addGroupMemberRequest
	var response interface{}

	err := util.ParseBody(r, &request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response = models.Response[string]{
			Data:         nil,
			Message:      "invalid request body",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Get the requester from the JWT context
	requester, ok := r.Context().Value(middleware.ContextKeyUsername).(string)
	if !ok || requester == "" {
		w.WriteHeader(http.StatusUnauthorized)
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not identify the requesting user",
			Error:        "unauthorized",
			StatusCode:   http.StatusUnauthorized,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Validate chat exists and is a group chat
	groupChat, chatErr := database.GetChat(request.ChatReference)
	if chatErr != nil || groupChat == nil {
		w.WriteHeader(http.StatusNotFound)
		response = models.Response[string]{
			Data:         nil,
			Message:      "chat not found",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	if groupChat.ChatType != database.ChatTypeGroup {
		w.WriteHeader(http.StatusBadRequest)
		response = models.Response[string]{
			Data:         nil,
			Message:      "members can only be removed from group chats",
			Error:        "",
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Validate requester is a member of this group
	requesterParticipant, _ := database.GetParticipant(requester, request.ChatReference)
	if requesterParticipant == nil {
		w.WriteHeader(http.StatusForbidden)
		response = models.Response[string]{
			Data:         nil,
			Message:      "you are not a member of this group chat",
			Error:        "",
			StatusCode:   http.StatusForbidden,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Validate the target user is actually a participant
	targetParticipant, _ := database.GetParticipant(request.Username, request.ChatReference)
	if targetParticipant == nil {
		w.WriteHeader(http.StatusNotFound)
		response = models.Response[string]{
			Data:         nil,
			Message:      request.Username + " is not a member of this group",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Delete participant
	_, pErr := database.DeleteParticipant(request.Username, request.ChatReference)
	if pErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not remove participant",
			Error:        pErr.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Deactivate the group chat reference on the user via gRPC
	targetUser, uErr := service.UserService.GetUser(
		context.Background(), &pb.UserRequest{UserId: request.Username},
	)
	if uErr == nil && targetUser != nil {
		_, grpcErr := service.UserService.RemoveUserConversation(
			context.Background(),
			&pb.RemoveUserConversationRequest{
				UserId:        targetUser.Id,
				ChatReference: request.ChatReference,
			},
		)
		if grpcErr != nil {
			fmt.Println("failed to deactivate group chat ref for", request.Username, ":", grpcErr)
		}
	}

	// Notify the removed user via their public socket
	groupRemoveStatus := "GROUP_REMOVED"
	activeUser := chat.ActiveSocketUsers[request.Username]
	if activeUser != nil && activeUser.Notify != nil {
		activeUser.Notify <- database.Message{
			MessageReference: uuid.NewString(),
			SenderUsername:   requester,
			ChatReference:    request.ChatReference,
			MessageStatus:    &groupRemoveStatus,
			SentTimestamp:    time.Now().Format(time.RFC3339),
		}
	}

	// Fetch updated participant list for the response
	participants, _ := database.GetParticipantsInChat(request.ChatReference)
	participantNames := []string{}
	for _, p := range participants {
		participantNames = append(participantNames, p.Username)
	}

	response = models.Response[models.NewGroupChat]{
		Data: &models.NewGroupChat{
			ChatReference: request.ChatReference,
			Name:          groupChat.Name,
			Participants:  participantNames,
		},
		Message:      request.Username + " removed from group",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}
