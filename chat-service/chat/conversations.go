package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/chat-service/database"
	"github.com/te6lim/go-chat/chat-service/models"
	"github.com/te6lim/go-chat/chat-service/service"

	pb "github.com/te6lim/go-chat-protos/userpb"
)

type Conversations struct {
	Conn   *websocket.Conn
	Chats  map[string]bool
	Notify chan database.Message
}

func SetupConversationsSocket(username string) {
	endpoint := fmt.Sprintf("/conversations/%s", username)
	service.Router.HandleFunc(endpoint, HandleConversations)
}

func HandleConversations(w http.ResponseWriter, r *http.Request) {
	conn, err := service.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	urlSegments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	username := urlSegments[1]
	user, errUser := service.UserService.GetUser(
		context.Background(),
		&pb.UserRequest{UserId: username},
	)
	if errUser != nil {
		w.WriteHeader(http.StatusNotFound)
		response := models.Response[string]{
			Data:         nil,
			Message:      "user does not exist",
			Error:        errUser.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	convsResp, convsErr := service.UserService.GetUserConversations(
		context.Background(),
		&pb.UserRequest{UserId: username},
	)
	if convsErr != nil {
		convsResp = &pb.UserConversationsResponse{}
	}

	socketUser, createErr := CreateSocketUser(user, convsResp.Conversations, AWAY)
	if createErr != nil {
		w.WriteHeader(http.StatusNotFound)
		response := models.Response[string]{
			Data:         nil,
			Message:      "user does not exist",
			Error:        createErr.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	socketUser.Conversations.Notify = make(chan database.Message)
	socketUser.Conversations.Conn = conn

	// Restore any hidden conversations that have unread messages so they
	// reappear in the user's conversation list on reconnect.
	chatsWithUnread, _ := database.GetChatsWithUnreadForUser(username)
	for _, c := range chatsWithUnread {
		if !socketUser.Chats[c.ChatReference] {
			_, _ = service.UserService.AddUserConversation(
				context.Background(),
				&pb.AddUserConversationRequest{
					UserId:        socketUser.UserId,
					ChatReference: c.ChatReference,
					ChatType:      string(c.ChatType),
				},
			)
			socketUser.Chats[c.ChatReference] = true
		}
	}

	defer func() {
		socketUser.Conn.Close()
		fmt.Println("conversations socket closed")
		LoggedOutUser <- socketUser
	}()

	socketUser.Activity = AWAY
	NewUserFromConversationsSetup <- socketUser
	go socketUser.ReadConversations()
	go sendUnacknowledgedMessagesForAllConversations(socketUser)
	socketUser.WriteConversations()
}

func (user *Socketuser) ReadConversations() {
	defer func() {
		user.Conn.Close()
		fmt.Println("conversations socket closed")
		LoggedOutUser <- user
	}()

	for {
		message := database.Message{}
		err := user.Conn.ReadJSON(&message)
		if err != nil {
			return
		}

		// Handle invite responses before any room forwarding.
		if message.MessageStatus != nil {
			switch *message.MessageStatus {
			case "ACCEPT_INVITE":
				acceptInvite(user, message)
				continue
			case "DECLINE_INVITE":
				declineInvite(user, message)
				continue
			}
		}

		var upToDateMessage *database.Message = &message
		var insertErr error = nil

		switch {
		case message.PresenceStatus == nil && message.MessageStatus == nil:
			upToDateMessage, insertErr = database.MaybeInsertAndReturnMostUpToDateMessage(&message)
		}

		room := Rooms[upToDateMessage.ChatReference]
		if room != nil {
			if insertErr != nil {
				fmt.Println(err)
				delete(Rooms, room.Id)
				return
			}
			room.ForwardedMessage <- *upToDateMessage
		} else {
			if insertErr != nil {
				return
			}
		}
	}
}

func acceptInvite(user *Socketuser, message database.Message) {
	chatRef := message.ChatReference
	invitee := user.Username

	myParticipant, _ := database.GetParticipant(invitee, chatRef)
	if myParticipant == nil || myParticipant.Status != database.ParticipantStatusPending {
		return
	}

	allParticipants, _ := database.GetParticipantsInChat(chatRef)
	initiatorUsername := ""
	for _, p := range allParticipants {
		if p.Username != invitee {
			initiatorUsername = p.Username
			break
		}
	}

	if _, err := database.UpdateParticipantStatus(invitee, chatRef, database.ParticipantStatusAccepted); err != nil {
		return
	}

	// Register the conversation in user-service for the invitee now that they accepted.
	inviteeUser, uErr := service.UserService.GetUser(context.Background(), &pb.UserRequest{UserId: invitee})
	if uErr == nil && inviteeUser != nil {
		initiatorUser, iErr := service.UserService.GetUser(context.Background(), &pb.UserRequest{UserId: initiatorUsername})
		if iErr == nil && initiatorUser != nil {
			_, _ = service.UserService.AddUserConversation(context.Background(), &pb.AddUserConversationRequest{
				UserId:        inviteeUser.Id,
				ChatReference: chatRef,
				ChatType:      "private",
				OtherUserId:   initiatorUser.Id,
			})
		}
	}

	// Track the chat in the invitee's in-memory map.
	user.Chats[chatRef] = true

	// Notify the initiator.
	acceptedStatus := "INVITE_ACCEPTED"
	activeInitiator := ActiveSocketUsers[initiatorUsername]
	if activeInitiator != nil && activeInitiator.Notify != nil {
		activeInitiator.Notify <- database.Message{
			MessageReference: uuid.NewString(),
			SenderUsername:   invitee,
			ChatReference:    chatRef,
			MessageStatus:    &acceptedStatus,
			SentTimestamp:    time.Now().Format(time.RFC3339),
		}
	}

	// Ensure the room socket endpoint is registered so both users can now connect.
	SetupRoomSocket(invitee, initiatorUsername, chatRef)
}

func declineInvite(user *Socketuser, message database.Message) {
	chatRef := message.ChatReference
	invitee := user.Username

	myParticipant, _ := database.GetParticipant(invitee, chatRef)
	if myParticipant == nil || myParticipant.Status != database.ParticipantStatusPending {
		return
	}

	allParticipants, _ := database.GetParticipantsInChat(chatRef)
	initiatorUsername := ""
	for _, p := range allParticipants {
		if p.Username != invitee {
			initiatorUsername = p.Username
			break
		}
	}

	for _, p := range allParticipants {
		database.DeleteParticipant(p.Username, chatRef)
	}
	database.DeleteChat(chatRef)

	if initiatorUsername == "" {
		return
	}

	initiatorUser, iErr := service.UserService.GetUser(context.Background(), &pb.UserRequest{UserId: initiatorUsername})
	if iErr != nil || initiatorUser == nil {
		return
	}

	_, _ = service.UserService.RemoveUserConversation(context.Background(), &pb.RemoveUserConversationRequest{
		UserId:        initiatorUser.Id,
		ChatReference: chatRef,
	})

	activeInitiator := ActiveSocketUsers[initiatorUsername]
	if activeInitiator == nil {
		return
	}

	if activeInitiator.Chats != nil {
		delete(activeInitiator.Chats, chatRef)
	}

	declinedStatus := "INVITE_DECLINED"
	if activeInitiator.Notify != nil {
		activeInitiator.Notify <- database.Message{
			MessageReference: uuid.NewString(),
			SenderUsername:   invitee,
			ChatReference:    chatRef,
			MessageStatus:    &declinedStatus,
			SentTimestamp:    time.Now().Format(time.RFC3339),
		}
	}
}

func (user *Socketuser) WriteConversations() {
	defer func() {
		user.Conn.Close()
		fmt.Println("conversations socket closed")
		LoggedOutUser <- user
	}()

	for {
		msg := <-user.Notify
		err := user.Conn.WriteJSON(msg)
		if err != nil {
			fmt.Println("Connection error: ", err)
			return
		}
	}
}

func (conversations *Conversations) sendUnacknowledgedMessages(chatRef string, user string) error {
	messages, err := database.GetMessagesWithIncompleteReceipts(chatRef, user)
	if err != nil {
		return err
	}

	for _, message := range messages {
		conversations.Notify <- message
	}
	return nil
}

func sendUnacknowledgedMessagesForAllConversations(socketUser *Socketuser) {
	for chatRef := range socketUser.Chats {
		go func() {
			socketUser.Conversations.sendUnacknowledgedMessages(chatRef, socketUser.Username)
		}()
	}
}
