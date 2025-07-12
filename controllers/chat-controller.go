package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/chat"
	"github.com/te6lim/go-chat/database"
	"github.com/te6lim/go-chat/requests"
	"github.com/te6lim/go-chat/responses"
	"github.com/te6lim/go-chat/utils"
)

type addChatReferenceRequest struct {
	User  string `json:"user"`
	Other string `json:"other"`
}

func HandleChat(templateHandler *utils.TemplateHandler) http.HandlerFunc {
	templateHandler.ParseFileOnce()
	return func(w http.ResponseWriter, r *http.Request) {
		me := mux.Vars(r)["username"]
		chatId := mux.Vars(r)["chatId"]
		var response interface{}
		userChat := database.GetChat(chatId)
		if userChat == nil {
			w.WriteHeader(http.StatusNotFound)
			response = responses.Response[string] {
				Data: nil,
				Message: "this chat does not exist",
				Error: "",
				StatusCode: http.StatusNotFound,
				IsSuccessful: false,
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
		var participants = database.GetParticipantsInChat(userChat.ChatReference)
		if len(participants) > 2 {
			w.WriteHeader(http.StatusForbidden)
			response = responses.Response[string] {
				Data: nil,
				Message: "too many participants",
				Error: "too many participants",
				StatusCode: http.StatusForbidden,
				IsSuccessful: false,
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
		var other *database.Participant
		var user *database.Participant
		for _, value := range participants {
			if value.Username != me {
				other = value
				break
			}
		}
		if other == nil {
			w.WriteHeader(http.StatusNotFound)
			response = responses.Response[string] {
				Data: nil,
				Message: "the other participant in this chat dooes not exist!",
				Error: "the other participant in this chat dooes not exist!",
				StatusCode: http.StatusNotFound,
				IsSuccessful: false,
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		for _, value := range participants {
			if value.Username == me {
				user = value
				break
			}
		}
		if user == nil {
			w.WriteHeader(http.StatusNotFound)
			response = responses.Response[string] {
				Data: nil,
				Message: "you are not a participant in this chat!",
				Error: "you are not a participant in this chat!",
				StatusCode: http.StatusNotFound,
				IsSuccessful: false,
			}
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
		chat.SetupSocketUser(me, other.Username, userChat.ChatReference)

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
	var response interface {}
	messages := database.GetAllUnacknowledgedMessages(chatRef, username)
	if messages == nil {
		response = responses.Response[string] {
			Data: nil,
			Message: "no unacknowledged messages",
			Error: "",
			StatusCode: http.StatusOK,
			IsSuccessful: true,
		}
		w.WriteHeader(http.StatusOK)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response = responses.Response[[]*database.Message] {
		Data: &messages,
		Message: "success",
		Error: "",
		StatusCode: http.StatusOK,
		IsSuccessful: true,
	}
	res, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func AcknowledgeMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var ackRequest = &requests.AckRequest{}
	var response interface{}
	err := utils.ParseBody(r, ackRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	messages := database.AcknowledgeMessages(
		ackRequest.ChatReference, ackRequest.Username, ackRequest.From, ackRequest.To,
	)
	if messages == nil {
		response = responses.Response[string] {
			Data: nil,
			Message: "no messages ackowledged",
			Error: "no message acknowledged",
			StatusCode: http.StatusOK,
			IsSuccessful: true,
		}
		w.WriteHeader(http.StatusOK)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response = responses.Response[[]*database.Message] {
		Data: &messages,
		Message: fmt.Sprintf("%v messages acknowledged", len(messages)),
		Error: "you are not a participant in this chat!",
		StatusCode: http.StatusOK,
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
		response = responses.Response[string] {
			Data: nil,
			Message: "couldn't add new message",
			Error: err.Error(),
			StatusCode: http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
	} else {
		response = responses.Response[database.Message] {
			Data: message,
			Message: "message added successfully",
			Error: "",
			StatusCode: http.StatusOK,
			IsSuccessful: true,
		}
		w.WriteHeader(http.StatusOK)
	}
	res, _ := json.Marshal(response)
	w.Write(res)
}

func DeleteMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	messageReference := r.URL.Query().Get("messageId")
	message := database.DeleteMessage(messageReference)
	res, err := json.Marshal(message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
}

func DeleteAllMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatId"]
	messages := database.DeleteAllMessages(chatRef)
	if messages == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := responses.Response[string] {
		Data: nil,
		Message: fmt.Sprintf("%v messages deleted ", len(messages)),
		Error: "",
		StatusCode: http.StatusOK,
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
	chat := database.GetChat(chatRef)
	var response interface{}
	if chat == nil {
		w.WriteHeader(http.StatusNotFound)
		response = responses.Response[string] {
			Data: nil,
			Message: chatRef + " not found",
			Error: "",
			StatusCode: http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	message := database.GetMessage(chat.ChatReference, messageRef)
	if message == nil {
		response = responses.Response[string] {
			Data: nil,
			Message: "message not found for " + chatRef,
			Error: "",
			StatusCode: http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		response = responses.Response[database.Message] {
			Data: message,
			Message: "",
			Error: "",
			StatusCode: http.StatusOK,
			IsSuccessful: true,
		}
		w.WriteHeader(http.StatusOK)
	}
	res, _ := json.Marshal(response)
	w.Write(res)
}

func GetAllMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatId"]
	chat := database.GetChat(chatRef)
	var response interface{}
	if chat == nil {
		response = responses.Response[string] {
			Data: nil,
			Message: chatRef + " not found",
			Error: "",
			StatusCode: http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.WriteHeader(http.StatusNotFound)
		w.Write(res)
		return
	}
	messages := database.GetAllMessages(chat.ChatReference)
	if messages == nil {
		response = responses.Response[string] {
			Data: nil,
			Message: "could not get messages",
			Error: "",
			StatusCode: http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		response = responses.Response[[]*database.Message] {
			Data: &messages,
			Message: "messages retrieved",
			Error: "",
			StatusCode: http.StatusOK,
			IsSuccessful: true,
		}
		w.WriteHeader(http.StatusOK)
	}
	res, _ := json.Marshal(response)
	w.Write(res)
}

func GetChatRefForUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := r.URL.Query().Get("user")
	other := r.URL.Query().Get("other")
	chatRef := database.GetChatRefFor(user, other)
	var response interface{}
	if chatRef == nil {
		response = responses.Response[string] {
			Data: nil,
			Message: "chat ref not found",
			Error: "",
			StatusCode: http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
	} else {
		response = responses.Response[string] {
			Data: chatRef,
			Message: "",
			Error: "",
			StatusCode: http.StatusOK,
			IsSuccessful: true,
		}
		w.WriteHeader(http.StatusOK)
	}
	res, _ := json.Marshal(response)
	w.Write(res)
}

func AddChatReference(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var refReq addChatReferenceRequest
	var response interface{}
	utils.ParseBody(r, &refReq)

	otheruser := database.GetUser(refReq.Other)
	user := database.GetUser(refReq.User)

	if user == nil || otheruser == nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	chatRef := database.GetChatRefFor(refReq.User, refReq.Other)
	if chatRef == nil {
		ref := uuid.NewString()
		_, err := database.InsertParticipant(database.Participant{
			Username:      refReq.User,
			ChatReference: ref,
		})
		if err != nil {
			response = responses.Response[string] {
				Data: nil,
				Message: "could not add participant to " + ref,
				Error: err.Error(),
				StatusCode: http.StatusBadRequest,
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
			response = responses.Response[string] {
				Data: nil,
				Message: "could not add participant to " + ref,
				Error: otherErr.Error(),
				StatusCode: http.StatusBadRequest,
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
			response = responses.Response[string] {
				Data: nil,
				Message: "",
				Error: chatErr.Error(),
				StatusCode: http.StatusBadRequest,
				IsSuccessful: false,
			}
			w.WriteHeader(http.StatusBadRequest)
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}

		user.AddConversation(otheruser.Id, chat.ChatReference)

		response = responses.Response[map[string]string] {
			Data: &map[string]string{
				"chatReference": chat.ChatReference,
			},
			Message: "",
			Error: "",
			StatusCode: http.StatusOK,
			IsSuccessful: true,
		}
		w.WriteHeader(http.StatusOK)
		res, _ := json.Marshal(response)
		w.Write(res)
	} else {
		user.AddConversation(otheruser.Id, *chatRef)

		response = responses.Response[responses.NewChat] {
			Data: &responses.NewChat{
				User:          refReq.User,
				Other:         refReq.Other,
				ChatReference: *chatRef,
			},
			Message: "",
			Error: "",
			StatusCode: http.StatusBadRequest,
			IsSuccessful: false,
		}

		w.WriteHeader(http.StatusOK)
		res, _ := json.Marshal(response)
		w.Write(res)
	}
}
