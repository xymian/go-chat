package chat

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/te6lim/go-chat/chat-service/database"

	pb "github.com/te6lim/go-chat-protos/userpb"
)

var (
	activeUsersMu    sync.RWMutex
	activeSocketUsers = make(map[string]*Socketuser)
)

var NewUserFromRoomSetup          chan *Socketuser = make(chan *Socketuser)
var NewUserFromConversationsSetup chan *Socketuser = make(chan *Socketuser)
var LoggedOutFromRoom             chan *Socketuser = make(chan *Socketuser)
var LoggedOutFromConversations    chan *Socketuser = make(chan *Socketuser)
var AwayUser                      chan *Socketuser = make(chan *Socketuser)

// GetActiveUser safely reads a user from the active users map.
func GetActiveUser(username string) *Socketuser {
	activeUsersMu.RLock()
	defer activeUsersMu.RUnlock()
	return activeSocketUsers[username]
}

func setActiveUser(username string, u *Socketuser) {
	activeUsersMu.Lock()
	defer activeUsersMu.Unlock()
	activeSocketUsers[username] = u
}

func removeActiveUser(username string) {
	activeUsersMu.Lock()
	defer activeUsersMu.Unlock()
	delete(activeSocketUsers, username)
}

func activeUsersCount() int {
	activeUsersMu.RLock()
	defer activeUsersMu.RUnlock()
	return len(activeSocketUsers)
}

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
			// Buffered so room.Run() is not blocked by a slow or exiting WriteMessages.
			IncomingMessage: make(chan database.Message, 32),
		},

		Conversations: conversations,
	}, nil
}

func ListenForActiveUsers() {
	for {
		select {
		case newUser := <-NewUserFromRoomSetup:
			activeUser := GetActiveUser(newUser.Username)
			if activeUser != nil {
				activeUser.PrivateChat = newUser.PrivateChat
				activeUser.Activity = newUser.Activity
			} else {
				setActiveUser(newUser.Username, newUser)
				activeUser = newUser
			}
			fmt.Println("number of users:", activeUsersCount())
			fmt.Println("New User", newUser.Username, " is ", activeUser.Activity.GetStatus())

		case newUser := <-NewUserFromConversationsSetup:
			activeUser := GetActiveUser(newUser.Username)
			if activeUser != nil {
				activeUser.Conversations = newUser.Conversations
				activeUser.Activity = newUser.Activity
			} else {
				setActiveUser(newUser.Username, newUser)
				activeUser = newUser
			}
			fmt.Println("number of users:", activeUsersCount())
			fmt.Println("New User", newUser.Username, " is ", activeUser.Activity.GetStatus())

		case user := <-LoggedOutFromRoom:
			active := GetActiveUser(user.Username)
			if active == nil {
				break
			}
			// Guard against a stale logout from a previous session being
			// processed after the user has already reconnected with a new
			// PrivateConn — only apply the cleanup if the conn matches.
			if active.PrivateConn != user.PrivateConn {
				break
			}
			active.PrivateConn = nil
			// If still connected via the conversations socket, stay in the
			// map as AWAY rather than removing the user entirely.
			if active.Conversations != nil && active.Conversations.Conn != nil {
				active.Activity = AWAY
			} else {
				removeActiveUser(user.Username)
			}
			fmt.Println("User", user.Username, " left room")

		case user := <-LoggedOutFromConversations:
			active := GetActiveUser(user.Username)
			if active == nil {
				break
			}
			// Guard against a stale logout from a previous session being
			// processed after the user has already reconnected with a new
			// conversations socket — only apply the cleanup if the
			// Conversations pointer matches the current session.
			if active.Conversations != user.Conversations {
				break
			}
			active.Conversations = nil
			// Only remove the user if the room socket is also gone.
			if active.PrivateConn == nil {
				removeActiveUser(user.Username)
			}
			fmt.Println("User", user.Username, " disconnected from conversations")
		}
	}
}
