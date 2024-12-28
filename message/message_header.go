package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"github.com/aang114/bitcoin-node/constants"
	"io"
)

type Checksum [checksumLength]byte

// Header of a Bitcoin p2p message (https://developer.bitcoin.org/reference/p2p_networking.html#message-headers)
type MessageHeader struct {
	// Magic value indicating message origin networking
	Magic uint32
	// ASCII string which identifies what message type is contained in the payload.
	Command CommandName
	// Number of bytes in payload
	Length uint32
	// First 4 bytes of SHA256(SHA256(payload)). If payload is empty, as in verack and “getaddr” messages, the Checksum is always 0x5df6e0e2 (SHA256(SHA256(<empty string>))).
	Checksum Checksum
}

func newMessageHeader(payload Payload) (MessageHeader, error) {
	encoded, err := payload.Encode()
	if err != nil {
		return MessageHeader{}, err
	}
	return MessageHeader{
		Magic:    constants.MainnetMagicValue,
		Command:  payload.CommandName(),
		Length:   uint32(len(encoded)),
		Checksum: checksum(encoded),
	}, nil
}

func (h *MessageHeader) encode() ([]byte, error) {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, h.Magic)
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(h.Command[:])
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, h.Length)
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(h.Checksum[:])
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeMessageHeader(r io.Reader) (*MessageHeader, error) {
	h := MessageHeader{}

	err := binary.Read(r, binary.LittleEndian, &h.Magic)
	if err != nil {
		return nil, err
	}
	_, err = io.ReadFull(r, h.Command[:])
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &h.Length)
	if err != nil {
		return nil, err
	}
	_, err = io.ReadFull(r, h.Checksum[:])
	if err != nil {
		return nil, err
	}

	return &h, nil
}

func checksum(encodedPayload []byte) Checksum {
	hash := sha256.Sum256(encodedPayload)
	hash = sha256.Sum256(hash[:])

	var checksum [checksumLength]byte
	copy(checksum[:], hash[0:checksumLength])

	return checksum
}
