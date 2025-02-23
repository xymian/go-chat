package socket

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

var Counter = 0

var Connections = make(map[string]*websocket.Conn)