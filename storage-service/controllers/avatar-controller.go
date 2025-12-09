package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/te6lim/go-chat/storage-service/models"
	"github.com/te6lim/go-chat/storage-service/service"

	storage_pb "github.com/te6lim/go-chat-protos/storage_pb"
)

const MaxUploadSize = 5 * 1024 * 1024

type StorageServer struct {
	service.LocalMediaStore
	storage_pb.UnimplementedStorageServiceServer
}

func (imageServer StorageServer) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
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
	if err := imageServer.Save(fileName, file); err != nil {
		response = models.Response[string]{
			Data:         nil,
			Message:      "storage error",
			Error:        err.Error(),
			StatusCode:   http.StatusInternalServerError,
			IsSuccessful: false,
		}
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(response)
		w.Write(res)
		return
	}

	// update user profile with avatar url

	response = models.Response[string]{
		Data:         nil,
		Message:      "avatar uploaded successfully",
		StatusCode:   http.StatusOK,
		IsSuccessful: false,
	}
	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(response)
	w.Write(res)
}
