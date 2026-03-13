package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/te6lim/go-chat/chat-service/chat"
	"github.com/te6lim/go-chat/chat-service/database"
	"github.com/te6lim/go-chat/chat-service/models"
	"github.com/te6lim/go-chat/chat-service/service"
	"github.com/te6lim/go-chat/chat-service/util"
	pb "github.com/te6lim/go-chat-protos/userpb"
)

type createGroupRequest struct {
	Name    string   `json:"name"`
	Creator string   `json:"creator"`
	Members []string `json:"members"`
}

type addGroupMemberRequest struct {
	Requester string `json:"requester"`
	Username  string `json:"username"`
}

type updateGroupNameRequest struct {
	Requester string `json:"requester"`
	Name      string `json:"name"`
}

type GroupInfo struct {
	Chat         database.Chat         `json:"chat"`
	Participants []database.Participant `json:"participants"`
}

// CreateGroupChat creates a new group chat.
// POST /group
// Body: { "name": "...", "creator": "alice", "members": ["bob", "charlie"] }
func CreateGroupChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req createGroupRequest
	var response interface{}

	if err := util.ParseBody(r, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Creator == "" {
		w.WriteHeader(http.StatusBadRequest)
		response = models.Response[string]{
			Data:         nil,
			Message:      "name and creator are required",
			Error:        "",
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Validate creator exists.
	_, creatorErr := service.UserService.GetUser(context.Background(), &pb.UserRequest{UserId: req.Creator})
	if creatorErr != nil {
		w.WriteHeader(http.StatusNotFound)
		response = models.Response[string]{
			Data:         nil,
			Message:      "creator " + req.Creator + " does not exist",
			Error:        creatorErr.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Validate all members exist.
	for _, member := range req.Members {
		if member == req.Creator {
			continue
		}
		_, err := service.UserService.GetUser(context.Background(), &pb.UserRequest{UserId: member})
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			response = models.Response[string]{
				Data:         nil,
				Message:      "member " + member + " does not exist",
				Error:        err.Error(),
				StatusCode:   http.StatusNotFound,
				IsSuccessful: false,
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
	}

	ref := uuid.NewString()
	name := req.Name
	creator := req.Creator

	newChat, chatErr := database.InsertChat(database.Chat{
		ChatReference: ref,
		IsGroup:       true,
		Name:          &name,
		CreatedBy:     &creator,
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

	// Insert creator as admin.
	_, pErr := database.InsertParticipant(database.Participant{
		Username:      req.Creator,
		ChatReference: ref,
		Role:          database.RoleAdmin,
	})
	if pErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response = models.Response[string]{
			Data:  nil,
			Error: pErr.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Insert remaining members.
	for _, member := range req.Members {
		if member == req.Creator {
			continue
		}
		_, mErr := database.InsertParticipant(database.Participant{
			Username:      member,
			ChatReference: ref,
			Role:          database.RoleMember,
		})
		if mErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response = models.Response[string]{
				Data:  nil,
				Error: mErr.Error(),
				StatusCode:   http.StatusInternalServerError,
				IsSuccessful: false,
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
	}

	participants, _ := database.GetParticipantsInChat(ref)
	response = models.Response[GroupInfo]{
		Data: &GroupInfo{
			Chat:         *newChat,
			Participants: participants,
		},
		Message:      "group created",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

// GetGroupInfo returns chat metadata and current member list.
// GET /group/{chatRef}
func GetGroupInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatRef"]
	var response interface{}

	groupChat, err := database.GetChat(chatRef)
	if err != nil || groupChat == nil || !groupChat.IsGroup {
		w.WriteHeader(http.StatusNotFound)
		response = models.Response[string]{
			Data:         nil,
			Message:      "group not found",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	participants, _ := database.GetParticipantsInChat(chatRef)
	response = models.Response[GroupInfo]{
		Data: &GroupInfo{
			Chat:         *groupChat,
			Participants: participants,
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

// UpdateGroupName updates the group's display name. Admin only.
// PUT /group/{chatRef}
// Body: { "requester": "alice", "name": "new name" }
func UpdateGroupName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatRef"]
	var req updateGroupNameRequest
	var response interface{}

	if err := util.ParseBody(r, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isAdmin, err := database.IsAdmin(req.Requester, chatRef)
	if err != nil || !isAdmin {
		w.WriteHeader(http.StatusForbidden)
		response = models.Response[string]{
			Data:         nil,
			Message:      "only admins can update the group name",
			Error:        "",
			StatusCode:   http.StatusForbidden,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	updated, updateErr := database.UpdateGroupName(chatRef, req.Name)
	if updateErr != nil || updated == nil {
		w.WriteHeader(http.StatusInternalServerError)
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not update group name",
			Error:        "",
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	response = models.Response[database.Chat]{
		Data:         updated,
		Message:      "group name updated",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

// AddGroupMember adds a new member to the group. Admin only.
// POST /group/{chatRef}/members
// Body: { "requester": "alice", "username": "dave" }
func AddGroupMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatRef"]
	var req addGroupMemberRequest
	var response interface{}

	if err := util.ParseBody(r, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isAdmin, err := database.IsAdmin(req.Requester, chatRef)
	if err != nil || !isAdmin {
		w.WriteHeader(http.StatusForbidden)
		response = models.Response[string]{
			Data:         nil,
			Message:      "only admins can add members",
			Error:        "",
			StatusCode:   http.StatusForbidden,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Ensure user exists.
	_, userErr := service.UserService.GetUser(context.Background(), &pb.UserRequest{UserId: req.Username})
	if userErr != nil {
		w.WriteHeader(http.StatusNotFound)
		response = models.Response[string]{
			Data:         nil,
			Message:      req.Username + " does not exist",
			Error:        userErr.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Check they are not already a member.
	existing, _ := database.GetParticipant(req.Username, chatRef)
	if existing != nil {
		w.WriteHeader(http.StatusConflict)
		response = models.Response[string]{
			Data:         nil,
			Message:      req.Username + " is already a member",
			Error:        "",
			StatusCode:   http.StatusConflict,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	participant, pErr := database.InsertParticipant(database.Participant{
		Username:      req.Username,
		ChatReference: chatRef,
		Role:          database.RoleMember,
	})
	if pErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not add member",
			Error:        pErr.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Refresh the room's member list if it is active.
	if room := chat.Rooms[chatRef]; room != nil {
		room.RefreshMembers()
	}

	response = models.Response[database.Participant]{
		Data:         participant,
		Message:      req.Username + " added to group",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

// RemoveGroupMember removes a member from the group.
// Admins can remove anyone; members can only remove themselves.
// DELETE /group/{chatRef}/members/{username}?requester=alice
func RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chatRef := mux.Vars(r)["chatRef"]
	targetUsername := mux.Vars(r)["username"]
	requester := r.URL.Query().Get("requester")
	var response interface{}

	if requester == "" {
		w.WriteHeader(http.StatusBadRequest)
		response = models.Response[string]{
			Data:         nil,
			Message:      "requester query param is required",
			Error:        "",
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Self-removal is always allowed; otherwise admin required.
	if requester != targetUsername {
		isAdmin, err := database.IsAdmin(requester, chatRef)
		if err != nil || !isAdmin {
			w.WriteHeader(http.StatusForbidden)
			response = models.Response[string]{
				Data:         nil,
				Message:      "only admins can remove other members",
				Error:        "",
				StatusCode:   http.StatusForbidden,
				IsSuccessful: false,
			}
			res, _ := json.Marshal(response)
			w.Write(res)
			return
		}
	}

	removed, removeErr := database.RemoveParticipant(targetUsername, chatRef)
	if removeErr != nil || removed == nil {
		w.WriteHeader(http.StatusNotFound)
		response = models.Response[string]{
			Data:         nil,
			Message:      targetUsername + " is not a member of this group",
			Error:        "",
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// Refresh the room's member list if it is active.
	if room := chat.Rooms[chatRef]; room != nil {
		room.RefreshMembers()
	}

	response = models.Response[database.Participant]{
		Data:         removed,
		Message:      targetUsername + " removed from group",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}
