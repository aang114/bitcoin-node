package message

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

// Services supported by a node (encoded as a bitfield) (https://en.bitcoin.it/wiki/Protocol_documentation#version)
type Services uint64

const (
	// This node is not a full node. It may not be able to provide any data except for the transactions it originates.
	Unnamed Services = 0
	// This node can be asked for full blocks instead of just headers.
	NodeNetwork Services = 1
	// This is a full node capable of responding to the getutxo protocol request. This is not supported by any currently-maintained Bitcoin node.
	NodeGetUtxo Services = 2
	// This is a full node capable and willing to handle bloom-filtered connections
	NodeBloom Services = 4
	// This is a full node that can be asked for blocks and transactions including witness data
	NodeWitness Services = 8
	// This is a full node that supports Xtreme Thinblocks. This is not supported by any currently-maintained Bitcoin node.
	NodeXThin Services = 16
	// See [BIP 0157](https://github.com/bitcoin/bips/blob/master/bip-0157.mediawiki)
	NodeCompactFilters Services = 64
	// This is the same as NODE_NETWORK but the node has at least the last 288 blocks (last 2 days)
	NodeNetworkLimited Services = 1024
)

// Network address of a node (https://en.bitcoin.it/wiki/Protocol_documentation#version)
type NetworkAddress struct {
	// Services supported by the node encoded as a bitfield
	Services Services
	// IP address of the node
	IpAddress net.IP
	// Port number of the node
	Port uint16
}

func NewNetworkAddress(services Services, ipAddress net.IP, port uint16) *NetworkAddress {
	return &NetworkAddress{
		Services:  services,
		IpAddress: ipAddress,
		Port:      port,
	}
}

func (n *NetworkAddress) encode() ([]byte, error) {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, n.Services)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.BigEndian, n.IpAddress.To16())
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.BigEndian, n.Port)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeNetworkAddress(r io.Reader) (*NetworkAddress, error) {
	n := NetworkAddress{}

	err := binary.Read(r, binary.LittleEndian, &n.Services)
	if err != nil {
		return nil, err
	}
	var ipAddressBytes = make([]byte, 16)
	_, err = io.ReadFull(r, ipAddressBytes)
	if err != nil {
		return nil, err
	}
	n.IpAddress = net.IP(ipAddressBytes)

	err = binary.Read(r, binary.BigEndian, &n.Port)
	if err != nil {
		return nil, err
	}
	return &n, nil
}
