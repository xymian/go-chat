package chat

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/chat-service/database"

	pb "github.com/te6lim/go-chat-protos/userpb"
)

var ActiveSocketUsers = make(map[string]*Socketuser)
var NewUserFromRoomSetup chan *Socketuser = make(chan *Socketuser)
var NewUserFromConversationsSetup chan *Socketuser = make(chan *Socketuser)
var LoggedOutUser chan *Socketuser = make(chan *Socketuser)
var AwayUser chan *Socketuser = make(chan *Socketuser)

type PrivateChat struct {
	PrivateConn     *websocket.Conn
	IncomingMessage chan database.Message
}

type Socketuser struct {
	PrivateChat
	*Conversations
	UserId   uint64
	Username string
	Activity PresenceStatus
}

func CreateSocketUser(user *pb.UserResponse, convs []*pb.UserConversation, activity PresenceStatus) (*Socketuser, error) {
	conversations := &Conversations{
		Chats: map[string]bool{},
	}

	for _, c := range convs {
		if c.Visible {
			conversations.Chats[c.ChatReference] = true
		}
	}

	return &Socketuser{
		UserId:   user.Id,
		Username: user.Username,
		Activity: activity,

		PrivateChat: PrivateChat{
			IncomingMessage: make(chan database.Message),
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
