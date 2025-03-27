package config

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var Router *mux.Router = mux.NewRouter()

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
