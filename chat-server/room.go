package chatserver

type Room interface {
	Leave()
	Join()
	Forward(channel chan []byte, message []byte)
}

type TwoUserRoom struct {
	ForwardedMessage chan[]byte
}

func (twoUserRoom *TwoUserRoom) Leave() {

}

func (twoUserRoom *TwoUserRoom) Join() {

}

func (twoUserRoom *TwoUserRoom) Forward( message []byte) {
	twoUserRoom.ForwardedMessage <- message
}

type MultiUserRoom struct {
}

func (multiUserRoom *MultiUserRoom) Leave() {

}

func (multiUserRoom *MultiUserRoom) Join() {

}

func (multiUserRoom *MultiUserRoom) Forward( message []byte) {
}
