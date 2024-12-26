package message

type GetAddrPayload struct{}

func (g GetAddrPayload) CommandName() CommandName {
	return GetAddrCommand
}

func (g GetAddrPayload) Encode() ([]byte, error) {
	return []byte{}, nil
}

func newGetAddrPayload() *GetAddrPayload {
	return &GetAddrPayload{}
}

func NewGetAddrMessage() (*Message, error) {
	payload := newGetAddrPayload()

	return newMessage(payload)
}
