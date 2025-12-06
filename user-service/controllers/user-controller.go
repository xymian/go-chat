package controllers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/xymian/go-chat/user-service/models"
	"github.com/xymian/go-chat/user-service/util"

	service "github.com/xymian/go-chat/user-service/service"

	pb "github.com/xymian/go-chat-protos/userpb"
)

const MaxUploadSize = 5 * 1024 * 1024

func UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	var response interface{}
	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "file too big or malformed request",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "invalid file key",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	defer file.Close()

	fileTypeBuff := make([]byte, 512)
	if _, err := file.Read(fileTypeBuff); err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "error reading file",
			Error:        err.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	filetype := http.DetectContentType(fileTypeBuff)

	var extention string
	switch filetype {
	case "image/jpeg":
		extention = ".jpg"
	case "image/png":
		extention = ".png"
	default:
		response = models.Response[string]{
			Data:         nil,
			Message:      "only JPEG and PNG allowed",
			Error:        err.Error(),
			StatusCode:   http.StatusUnsupportedMediaType,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusUnsupportedMediaType)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "server error processing file",
			Error:        err.Error(),
			StatusCode:   http.StatusUnsupportedMediaType,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusUnsupportedMediaType)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	randomBytesForFileNaming := make([]byte, 16)
	rand.Read(randomBytesForFileNaming)
	fileName := fmt.Sprintf("u_%s_%s%s", "CURRENT_USER_ID", hex.EncodeToString(randomBytesForFileNaming), extention)

	// make call to save file

	response = models.Response[string]{
		Data:         nil,
		Message:      "avatar uploaded successfully",
		StatusCode:   http.StatusOK,
		IsSuccessful: false,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
	return
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	username := mux.Vars(r)["username"]
	_, err := service.UserServer.GetUser(context.Background(), &pb.UserRequest{
		UserId: username,
	})

	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "user does not exist",
			Error:        err.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	//TODO: Fix this
	response = models.Response[pb.UserResponse]{
		Data:         nil,
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)

	res, _ := json.Marshal(response)
	w.Write(res)
}

func GetInteractions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	username := mux.Vars(r)["username"]
	var response interface{}
	user, err := service.UserServer.GetUser(context.Background(), &pb.UserRequest{
		UserId: username,
	})
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
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
	conversationMap := map[int64]string{}
	if len(user.ChatReferences) != 0 {
		err := json.Unmarshal([]byte(user.ChatReferences), &conversationMap)
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
	}

	userIds := []int64{}
	for k := range conversationMap {
		userIds = append(userIds, k)
	}
	userList, err := service.UserServer.GetUsers(context.Background(), nil)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
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
	userChatInfo := []*models.UserChatInfo{}
	for _, u := range userList.Users {
		userChatInfo = append(userChatInfo, &models.UserChatInfo{
			Username:        u.Username,
			DisplayImageUrl: "",
			ChatReference:   conversationMap[int64(u.Id)],
		})
	}

	response = models.Response[[]*models.UserChatInfo]{
		Data:         &userChatInfo,
		Message:      "success",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}

	res, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	_, err := service.UserServer.GetUsers(context.Background(), nil)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "couldn't get all users",
			Error:        err.Error(),
			StatusCode:   http.StatusNotFound,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusNotFound)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	response = models.Response[string]{
		Data:         nil,
		Message:      "",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: false,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}

func InsertUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response interface{}
	user := &pb.UserResponse{}
	err := util.ParseBody(r, &user)
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "bad request",
			Error:        err.Error(),
			StatusCode:   http.StatusBadRequest,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	_, uErr := service.UserServer.InsertUser(context.Background(), user)
	if uErr != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "could not insert user",
			Error:        uErr.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}
	response = models.Response[string]{
		Data:         nil,
		Message:      "new user added",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: true,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	var response interface{}
	user, err := service.UserServer.DeleteUser(context.Background(), &pb.UserRequest{UserId: username})
	if err != nil {
		response = models.Response[string]{
			Data:         nil,
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
	response = models.Response[string]{
		Data:         &user.Username,
		Message:      "user has been deleted",
		Error:        "",
		StatusCode:   http.StatusOK,
		IsSuccessful: false,
	}
	res, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}
