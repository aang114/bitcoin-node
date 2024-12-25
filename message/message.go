package message

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

const (
	commandNameLength        = 12
	checksumLength           = 4
	maxPayloadSize    uint32 = 32 * 1024 * 1024
)

var (
	ErrPayloadTooBig        = errors.New("payload too big")
	ErrInvalidChecksum      = errors.New("invalid Checksum")
	ErrInvalidPayloadLength = errors.New("invalid Payload length")
)

var (
	VersionCommand = CommandName{'v', 'e', 'r', 's', 'i', 'o', 'n'}
	VerackCommand  = CommandName{'v', 'e', 'r', 'a', 'c', 'k'}
)

type CommandName [commandNameLength]byte

type Payload interface {
	CommandName() CommandName
	Encode() ([]byte, error)
}

// A Bitcoin p2p message (https://en.bitcoin.it/wiki/Protocol_documentation#Message_structure)
type Message struct {
	Header  MessageHeader
	Payload Payload
}

func newMessage(payload Payload) (*Message, error) {
	header, err := newMessageHeader(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		Header:  header,
		Payload: payload,
	}, nil
}

func (f *Message) Encode() ([]byte, error) {
	encodedHeader, err := f.Header.encode()
	if err != nil {
		return nil, err
	}
	encodedMessage, err := f.Payload.Encode()
	if err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)
	_, err = buffer.Write(encodedHeader)
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(encodedMessage)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func DecodeMessage(r io.Reader) (*Message, error) {
	header, err := decodeMessageHeader(r)
	if err != nil {
		return nil, err
	}
	if header.Length > maxPayloadSize {
		return nil, ErrPayloadTooBig
	}

	encodedPayload := make([]byte, header.Length)
	_, err = io.ReadFull(r, encodedPayload)
	if err != nil {
		return nil, err
	}
	if header.Checksum != checksum(encodedPayload) {
		return nil, ErrInvalidChecksum
	}

	var payload Payload
	switch header.Command {
	case VersionCommand:
		payload, err = decodeVersionPayload(bytes.NewReader(encodedPayload))
		if err != nil {
			return nil, err
		}
	case VerackCommand:
		if len(encodedPayload) != 0 {
			return nil, ErrInvalidPayloadLength
		}
		payload = &VerackPayload{}
	default:
		//  TODO error
		return nil, errors.New(fmt.Sprintf("unknown commnad name: %s", header.Command))
	}

	return &Message{
		Header:  *header,
		Payload: payload,
	}, nil
}
