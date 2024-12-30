package message

import (
	"bytes"
	"encoding/binary"
	"io"
)

// The “Version” message provides information about the transmitting node to the receiving node at the beginning of a connection.
// Until both peers have exchanged “Version” messages, no other messages will be accepted.
//
// Source: https://developer.bitcoin.org/reference/p2p_networking.html#version
type VersionPayload struct {
	// Highest protocol Version understood by the transmitting node
	Version int32
	// Services supported by the transmitting node encoded as a bitfield
	Services Services
	// Current Unix time according to the transmitting node’s clock
	Timestamp int64
	// Receiving node as perceived by the transmitting node
	ReceivingNode NetworkAddress
	// Transmitting Node
	TransmittingNode NetworkAddress
	// Random Nonce which can help a node detect a connection to itself
	//
	// If the Nonce is 0, the Nonce field is ignored.
	// If the Nonce is anything else, a node should terminate the connection on receipt of a “Version” message with a Nonce it previously sent.
	Nonce uint64
	// User agent as defined by [BIP14](https://github.com/bitcoin/bips/blob/master/bip-0014.mediawiki)
	UserAgent string
	// Height of the transmitting node’s best block
	StartHeight int32
	// Whether the remote peer should announce relayed transactions or not, see (BIP 0037)[https://github.com/bitcoin/bips/blob/master/bip-0037.mediawiki]
	Relay bool
}

func newVersionPayload(
	version int32,
	services Services,
	timestamp int64,
	receivingNode NetworkAddress,
	transmittingNode NetworkAddress,
	nonce uint64,
	userAgent string,
	startHeight int32,
	relay bool,
) *VersionPayload {
	return &VersionPayload{
		Version:          version,
		Services:         services,
		Timestamp:        timestamp,
		ReceivingNode:    receivingNode,
		TransmittingNode: transmittingNode,
		Nonce:            nonce,
		UserAgent:        userAgent,
		StartHeight:      startHeight,
		Relay:            relay,
	}
}

func NewVersionMessage(
	version int32,
	services Services,
	timestamp int64,
	receivingNode NetworkAddress,
	transmittingNode NetworkAddress,
	nonce uint64,
	userAgent string,
	startHeight int32,
	relay bool,
) (*Message, error) {
	payload := newVersionPayload(
		version,
		services,
		timestamp,
		receivingNode,
		transmittingNode,
		nonce,
		userAgent,
		startHeight,
		relay)

	return newMessage(payload)
}

func (v VersionPayload) CommandName() CommandName {
	return VersionCommand
}

func (v VersionPayload) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, v.Version)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, v.Services)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, v.Timestamp)
	if err != nil {
		return nil, err
	}

	encodedReceivingNode, err := v.ReceivingNode.encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(encodedReceivingNode)
	if err != nil {
		return nil, err
	}
	encodedTransmittingNode, err := v.TransmittingNode.encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(encodedTransmittingNode)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, v.Nonce)
	if err != nil {
		return nil, err
	}
	userAgentLengthEncoded, err := VarInt(len(v.UserAgent)).Encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(userAgentLengthEncoded)
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write([]byte(v.UserAgent))
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, v.StartHeight)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, v.Relay)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeVersionPayload(r io.Reader) (*VersionPayload, error) {
	v := VersionPayload{}

	err := binary.Read(r, binary.LittleEndian, &v.Version)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &v.Services)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &v.Timestamp)
	if err != nil {
		return nil, err
	}

	receivingNode, err := decodeNetworkAddress(r)
	if err != nil {
		return nil, err
	}
	v.ReceivingNode = *receivingNode
	transmittingNode, err := decodeNetworkAddress(r)
	if err != nil {
		return nil, err
	}
	v.TransmittingNode = *transmittingNode

	err = binary.Read(r, binary.LittleEndian, &v.Nonce)
	if err != nil {
		return nil, err
	}

	userAgentLen, err := DecodeVarInt(r)
	if err != nil {
		return nil, err
	}
	userAgentBytes := make([]byte, userAgentLen)
	_, err = io.ReadFull(r, userAgentBytes)
	if err != nil {
		return nil, err
	}
	// TODO - Make sure string is UTF-8
	v.UserAgent = string(userAgentBytes)

	err = binary.Read(r, binary.LittleEndian, &v.StartHeight)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &v.Relay)
	if err != nil {
		return nil, err
	}

	return &v, nil
}
