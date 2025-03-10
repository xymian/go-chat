package config

import chatserver "github.com/te6lim/go-chat/chat-server"

var Session chan *chatserver.UserSession = make(chan *chatserver.UserSession)
