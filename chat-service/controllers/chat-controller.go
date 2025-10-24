package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/chat-service/chat"
	"github.com/te6lim/go-chat/chat-service/database"
	"github.com/te6lim/go-chat/chat-service/models"
	"github.com/te6lim/go-chat/chat-service/service"
	"github.com/te6lim/go-chat/chat-service/util"
	pb "github.com/xymian/go-chat-protos/userpb"
)

type addChatReferenceRequest struct {
	User  string `json:"user"`
	Other string `json:"other"`
}

func MarkMessagesAsDelivered(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	deliverMessage := &models.DeliverMessages{}
	var response models.Response[[]database.Message]
	err := util.ParseBody(r, deliverMessage)

	if err != nil {
		response = models.Response[[]database.Message]{
			Data:         nil,
			Message:      "bad request",
			Error:        "request parse error",
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	messages, err := database.MarkMessagesAsDelivered(*deliverMessage)
	if err != nil {
		response = models.Response[[]database.Message]{
			Data:         &messages,
			Message:      "",
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
		Message:      "messages marked as delivered",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func SetupPublicSocket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := mux.Vars(r)["username"]
	chat.SetUpPublicSocket(username)
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

func GetUnacknowledgedMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatId"]
	username := mux.Vars(r)["username"]
	var response interface{}
	messages, err := database.GetAllUnacknowledgedMessages(chatRef, username)
	if err != nil {
		response = models.Response[[]database.Message]{
			Data:         &messages,
			Message:      "no unacknowledged messages",
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
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	res, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func AcknowledgeMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var ackRequest = &models.AckRequest{}
	var response interface{}
	err := util.ParseBody(r, ackRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	messages, err := database.AcknowledgeMessages(
		ackRequest.ChatReference, ackRequest.Username, ackRequest.From, ackRequest.To,
	)
	if err != nil {
		response = models.Response[[]database.Message]{
			Data:         &messages,
			Message:      "",
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
		Message:      fmt.Sprintf("%v messages acknowledged", len(messages)),
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	res, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
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

			_, err := service.UserService.AddConversation(context.Background(), &pb.AddChatRequest{
				User: &pb.UserResponse{
					Id:             uint64(user.Id),
					Username:       user.Username,
					PasswordHash:   user.PasswordHash,
					ChatReferences: user.ChatReferences,
					CreatedAt:      user.CreatedAt,
					UpdatedAt:      user.UpdatedAt,
				},
				Chat: &pb.Chat{
					Username: otheruser.Username,
					ChatRef:  chat.ChatReference,
				},
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
			_, err := service.UserService.AddConversation(context.Background(), &pb.AddChatRequest{
				User: &pb.UserResponse{
					Id:             uint64(user.Id),
					Username:       user.Username,
					PasswordHash:   user.PasswordHash,
					ChatReferences: user.ChatReferences,
					CreatedAt:      user.CreatedAt,
					UpdatedAt:      user.UpdatedAt,
				},
				Chat: &pb.Chat{
					Username: otheruser.Username,
					ChatRef:  *chatRef,
				},
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
