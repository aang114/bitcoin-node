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

type ErrUnknownCommandName struct {
	Command CommandName
}

func (e *ErrUnknownCommandName) Error() string {
	return fmt.Sprintf("unknown command name: %s", e.Command)
}

var (
	VersionCommand    = CommandName{'v', 'e', 'r', 's', 'i', 'o', 'n'}
	VerackCommand     = CommandName{'v', 'e', 'r', 'a', 'c', 'k'}
	WtxidRelayCommand = CommandName{'w', 't', 'x', 'i', 'd', 'r', 'e', 'l', 'a', 'y'}
	SendAddrV2Command = CommandName{'s', 'e', 'n', 'd', 'a', 'd', 'd', 'r', 'v', '2'}
	GetAddrCommand    = CommandName{'g', 'e', 't', 'a', 'd', 'd', 'r'}
	AddrCommand       = CommandName{'a', 'd', 'd', 'r'}
	GetBlocksCommand  = CommandName{'g', 'e', 't', 'b', 'l', 'o', 'c', 'k', 's'}
	InvCommand        = CommandName{'i', 'n', 'v'}
	GetDataCommand    = CommandName{'g', 'e', 't', 'd', 'a', 't', 'a'}
	BlockCommand      = CommandName{'b', 'l', 'o', 'c', 'k'}
	TxCommand         = CommandName{'t', 'x'}
	PingCommand       = CommandName{'p', 'i', 'n', 'g'}
	PongCommand       = CommandName{'p', 'o', 'n', 'g'}
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
	case VerackCommand:
		if len(encodedPayload) != 0 {
			return nil, ErrInvalidPayloadLength
		}
		payload = &VerackPayload{}
	case WtxidRelayCommand:
		if len(encodedPayload) != 0 {
			return nil, ErrInvalidPayloadLength
		}
		payload = &WtxidRelayPayload{}
	case SendAddrV2Command:
		if len(encodedPayload) != 0 {
			return nil, ErrInvalidPayloadLength
		}
		payload = &SendAddrV2Payload{}
	case AddrCommand:
		payload, err = decodeAddrPayload(bytes.NewReader(encodedPayload))
	case GetAddrCommand:
		if len(encodedPayload) != 0 {
			return nil, ErrInvalidPayloadLength
		}
		payload = &GetAddrPayload{}
	case GetBlocksCommand:
		payload, err = decodeGetBlocksPayload(bytes.NewReader(encodedPayload))
	case InvCommand:
		payload, err = decodeInvPayload(bytes.NewReader(encodedPayload))
	case GetDataCommand:
		payload, err = decodeGetDataPayload(bytes.NewReader(encodedPayload))
	case TxCommand:
		payload, err = decodeTxPayload(bytes.NewReader(encodedPayload))
	case BlockCommand:
		payload, err = decodeBlockPayload(bytes.NewReader(encodedPayload))
	case PingCommand:
		payload, err = decodePingPayload(bytes.NewReader(encodedPayload))
	case PongCommand:
		payload, err = decodePongPayload(bytes.NewReader(encodedPayload))
	default:
		return nil, &ErrUnknownCommandName{Command: header.Command}
	}
	if err != nil {
		return nil, err
	}

	return &Message{
		Header:  *header,
		Payload: payload,
	}, nil
}
