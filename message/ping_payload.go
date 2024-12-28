package message

import (
	"bytes"
	"encoding/binary"
	"io"
)

// https://en.bitcoin.it/wiki/Protocol_documentation#ping
type PingPayload struct {
	Nonce uint64
}

func (p *PingPayload) CommandName() CommandName {
	return PingCommand
}

func (p *PingPayload) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, p.Nonce)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decodePingPayload(r io.Reader) (*PingPayload, error) {
	p := PingPayload{}
	err := binary.Read(r, binary.LittleEndian, &p.Nonce)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func newPingPayload(nonce uint64) *PingPayload {
	return &PingPayload{
		Nonce: nonce,
	}
}

func NewPingMessage(nonce uint64) (*Message, error) {
	payload := newPingPayload(nonce)
	return newMessage(payload)
}
