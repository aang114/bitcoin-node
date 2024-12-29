package message

type WtxidRelayPayload struct{}

func (w *WtxidRelayPayload) CommandName() CommandName {
	return WtxidRelayCommand
}

func (w *WtxidRelayPayload) Encode() ([]byte, error) {
	return []byte{}, nil
}

func newWtxidRelayPayload() *WtxidRelayPayload {
	return &WtxidRelayPayload{}
}

func NewWtxidRelayMessage() (*Message, error) {
	payload := newWtxidRelayPayload()
	return newMessage(payload)
}
