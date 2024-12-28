package message

import (
	"bytes"
	"encoding/binary"
	"io"
)

type PongPayload struct {
	Nonce uint64
}

func (p *PongPayload) CommandName() CommandName {
	return PongCommand
}

func (p *PongPayload) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, p.Nonce)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decodePongPayload(r io.Reader) (*PongPayload, error) {
	p := PongPayload{}
	err := binary.Read(r, binary.LittleEndian, &p.Nonce)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func newPongPayload(nonce uint64) *PongPayload {
	return &PongPayload{
		Nonce: nonce,
	}
}

func NewPongMessage(nonce uint64) (*Message, error) {
	payload := newPongPayload(nonce)
	return newMessage(payload)
}
