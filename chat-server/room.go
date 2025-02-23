package chatserver

import (
	"errors"
)

type Room interface {
	Leave()
	Join(user *User) error
	Forward(channel chan []byte, message []byte)
}

type TwoUserRoom struct {
	leave            chan *User
	join             chan *User
	participants     map[*User]bool
	ForwardedMessage chan []byte
}

type TwoUserRoomPayload struct {
	Room *TwoUserRoom
	Id   string
}

var PrivateRooms = make(map[string]*TwoUserRoom)

var NewRoom chan *TwoUserRoomPayload

func (twoUserRoom *TwoUserRoom) Leave() {
	twoUserRoom.leave <- nil
}

func (twoUserRoom *TwoUserRoom) Join(user *User) error {
	if twoUserRoom.join == nil {
		twoUserRoom.join <- user
	} else {
		return errors.New("Room is full")
	}
	return nil
}

func (twoUserRoom *TwoUserRoom) Forward(message []byte) {
	twoUserRoom.ForwardedMessage <- message
}

type MultiUserRoom struct {
}

func (multiUserRoom *MultiUserRoom) Leave() {

}

func (multiUserRoom *MultiUserRoom) Join(user *User) error {
	return nil
}

func (multiUserRoom *MultiUserRoom) Forward(message []byte) {
}

func CreateTwoUserRoom(id string) *TwoUserRoom {
	return &TwoUserRoom{
		leave:            make(chan *User),
		join:             make(chan *User),
		participants:     make(map[*User]bool),
		ForwardedMessage: make(chan []byte),
	}
}

func (twoUserRoom *TwoUserRoom) run() {
	for {
		select {
		case user := <-twoUserRoom.join:
			if user != nil {
				twoUserRoom.participants[user] = true
			}

		case user := <-twoUserRoom.leave:
			if user != nil {
				twoUserRoom.participants[user] = false
				delete(twoUserRoom.participants, user)
				close(user.Message)
			}

		case message := <-twoUserRoom.ForwardedMessage:
			for user := range twoUserRoom.participants {
				user.Message <- message
			}
		}
	}
}

func (user *User) ListenForAddedRooms() {
	for {
		select {
		case newRoom := <-NewRoom:
			if user.PrivateRooms[newRoom.Id] != nil {
				user.PrivateRooms[newRoom.Id] = newRoom.Room
			}
		}
	}
}
