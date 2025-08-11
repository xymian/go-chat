package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/chat"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/models"
	"github.com/te6lim/go-chat/utils"
)

type addChatReferenceRequest struct {
	User  string `json:"user"`
	Other string `json:"other"`
}

func MarkMessagesAsDelivered(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	deliverMessage := &models.DeliverMessages{}
	var response models.Response[[]database.Message]
	err := utils.ParseBody(r, deliverMessage)

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

func SetupUniqueIntractionsSocket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := mux.Vars(r)["username"]
	chat.SetUpInteractionsSocket(
		username,
		func(isConnected bool) {
			if isConnected {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		},
	)
}

func SetupUniqueSocket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newChat = &models.NewChat{}
	utils.ParseBody(r, newChat)
	me := newChat.User
	other := newChat.Other
	chatId := newChat.ChatReference
	chat.SetupSocketUser(
		me, other, chatId,
		func(isConnected bool) {
			if isConnected {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		},
	)
	w.WriteHeader(http.StatusOK)
}

func HandleChat(templateHandler *utils.TemplateHandler) http.HandlerFunc {
	templateHandler.ParseFileOnce()
	return func(w http.ResponseWriter, r *http.Request) {
		me := mux.Vars(r)["username"]
		chatId := mux.Vars(r)["chatId"]
		var response interface{}
		userChat, err := database.GetChat(chatId)
		if err != nil {
			response = models.Response[string]{
				Data:         nil,
				Message:      "this chat does not exist",
				Error:        "",
				StatusCode:   http.StatusNotFound,
				IsSuccessful: false,
			}
			w.WriteHeader(http.StatusNotFound)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		var participants, pErr = database.GetParticipantsInChat(userChat.ChatReference)
		if pErr != nil {
			response = models.Response[string]{
				Data:         nil,
				Message:      "",
				Error:        pErr.Error(),
				StatusCode:   http.StatusNotFound,
				IsSuccessful: false,
			}
			w.WriteHeader(http.StatusNotFound)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
		if len(participants) > 2 {
			response = models.Response[string]{
				Data:         nil,
				Message:      "too many participants",
				Error:        "too many participants",
				StatusCode:   http.StatusForbidden,
				IsSuccessful: false,
			}
			w.WriteHeader(http.StatusForbidden)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
		var other *database.Participant
		var user *database.Participant
		for _, value := range participants {
			if value.Username != me {
				other = &value
				break
			}
		}
		if other == nil {
			response = models.Response[string]{
				Data:         nil,
				Message:      "the other participant in this chat dooes not exist!",
				Error:        "the other participant in this chat dooes not exist!",
				StatusCode:   http.StatusNotFound,
				IsSuccessful: false,
			}
			w.WriteHeader(http.StatusNotFound)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		for _, value := range participants {
			if value.Username == me {
				user = &value
				break
			}
		}
		if user == nil {

			response = models.Response[string]{
				Data:         nil,
				Message:      "you are not a participant in this chat!",
				Error:        "you are not a participant in this chat!",
				StatusCode:   http.StatusNotFound,
				IsSuccessful: false,
			}
			w.WriteHeader(http.StatusNotFound)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		data := map[string]interface{}{
			"Host":   r.Host,
			"ChatId": userChat.ChatReference,
			"Me":     user.Username,
			"Other":  other.Username,
		}
		chat.SetupSocketUser(
			me, other.Username, userChat.ChatReference,
			func(isConnected bool) {
				if (isConnected) {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
		)

		templateHandler.Template.Execute(w, data)
	}
}

func HandleNewChat(templatehandler *utils.TemplateHandler) http.HandlerFunc {
	templatehandler.ParseFileOnce()
	return func(w http.ResponseWriter, r *http.Request) {
		me := mux.Vars(r)["username"]
		data := map[string]interface{}{
			"Host": r.Host,
			"Me":   me,
		}
		templatehandler.Template.Execute(w, data)
	}
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
	err := utils.ParseBody(r, ackRequest)
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
	e := utils.ParseBody(r, &message)
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
	var refReq addChatReferenceRequest
	var response interface{}
	utils.ParseBody(r, &refReq)

	otheruser, oErr := database.GetUser(refReq.Other)
	user, uErr := database.GetUser(refReq.User)

	if uErr != nil {
		w.WriteHeader(http.StatusNotFound)
		response := models.Response[string]{
			Data:         nil,
			Message:      "username " + refReq.User + " does not exist",
			Error:        uErr.Error(),
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
			Message:      "username " + refReq.Other + " does not exist",
			Error:        oErr.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	chatRef, cErr := database.GetChatRefFor(refReq.User, refReq.Other)
	if cErr == nil {
		if chatRef == nil {
			ref := uuid.NewString()
			_, pErr := database.InsertParticipant(database.Participant{
				Username:      refReq.User,
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
				Username:      refReq.Other,
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

			_, err := user.AddConversation(otheruser.Id, chat.ChatReference)
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
					User:          refReq.User,
					Other:         refReq.Other,
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
			_, err := user.AddConversation(otheruser.Id, *chatRef)

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
					User:          refReq.User,
					Other:         refReq.Other,
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
