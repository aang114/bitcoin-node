package message

type VerackPayload struct{}

func (v *VerackPayload) CommandName() CommandName {
	return VerackCommand
}

func (v *VerackPayload) Encode() ([]byte, error) {
	return []byte{}, nil
}

func newVerackPayload() *VerackPayload {
	return &VerackPayload{}
}

func NewVerackMessage() (*Message, error) {
	payload := newVerackPayload()
	return newMessage(payload)
}
