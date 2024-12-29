package message

type SendAddrV2Payload struct{}

func (s *SendAddrV2Payload) CommandName() CommandName {
	return SendAddrV2Command
}

func (s *SendAddrV2Payload) Encode() ([]byte, error) {
	return []byte{}, nil
}

func newSendAddrV2Payload() *SendAddrV2Payload {
	return &SendAddrV2Payload{}
}

func NewSendAddrV2Message() (*Message, error) {
	payload := newSendAddrV2Payload()
	return newMessage(payload)
}
