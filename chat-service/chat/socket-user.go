package chat

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/chat-service/database"

	pb "github.com/xymian/go-chat-protos/userpb"
)

var ActiveSocketUsers = make(map[string]*Socketuser)
var NewUserFromRoomSetup chan *Socketuser = make(chan *Socketuser)
var NewUserFromConversationsSetup chan *Socketuser = make(chan *Socketuser)
var LoggedOutUser chan *Socketuser = make(chan *Socketuser)
var AwayUser chan *Socketuser = make(chan *Socketuser)

type PrivateChat struct {
	PrivateConn    *websocket.Conn
	ReceiveMessage chan database.Message
}

type Socketuser struct {
	PrivateChat
	*Conversations
	Username string
	Activity PresenceStatus
}

func CreateSocketUser(user *pb.UserResponse, activity PresenceStatus) (*Socketuser, error) {
	conversations := &Conversations{
		Chats: map[string]bool{},
	}
	conversationMap := map[int64]string{}
	err := json.Unmarshal([]byte(user.ChatReferences), &conversationMap)
	if err != nil {
		return nil, err
	}
	for _, chatRef := range conversationMap {
		conversations.Chats[chatRef] = true
	}
	return &Socketuser{
		Username: user.Username,
		Activity: activity,

		PrivateChat: PrivateChat{
			ReceiveMessage: make(chan database.Message),
		},

		Conversations: conversations,
	}, nil
}

func ListenForActiveUsers() {
	for {
		select {
		case newUser := <-NewUserFromRoomSetup:
			activeUser := ActiveSocketUsers[newUser.Username]
			if activeUser != nil {
				activeUser.PrivateChat = newUser.PrivateChat
				activeUser.Activity = newUser.Activity
			} else {
				ActiveSocketUsers[newUser.Username] = newUser
				activeUser = newUser
			}

			fmt.Println("number of users: ", len(ActiveSocketUsers))
			fmt.Println("New User", newUser.Username, " is ", activeUser.Activity.GetStatus())

		case newUser := <-NewUserFromConversationsSetup:
			activeUser := ActiveSocketUsers[newUser.Username]
			if activeUser != nil {
				activeUser.Conversations = newUser.Conversations
				activeUser.Activity = newUser.Activity
			} else {
				ActiveSocketUsers[newUser.Username] = newUser
				activeUser = newUser
			}
			fmt.Println("number of users: ", len(ActiveSocketUsers))
			fmt.Println("New User", newUser.Username, " is ", activeUser.Activity.GetStatus())

		case loggedOutUser := <-LoggedOutUser:
			delete(ActiveSocketUsers, loggedOutUser.Username)
			fmt.Println("User", loggedOutUser.Username, " logged out")
		}
	}
}
