package message

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// https://en.bitcoin.it/wiki/Protocol_documentation#addr
const maxAddrCount = 1000

type Address struct {
	Timestamp      uint32
	NetworkAddress NetworkAddress
}

func NewAddress(timestamp uint32, networkAddress NetworkAddress) *Address {
	return &Address{
		Timestamp:      timestamp,
		NetworkAddress: networkAddress,
	}
}

type AddrPayload struct {
	AddressList []Address
}

func newAddrPayload(addressList []Address) *AddrPayload {
	return &AddrPayload{
		AddressList: addressList,
	}
}

func NewAddrMessage(addressList []Address) (*Message, error) {
	payload := newAddrPayload(addressList)
	return newMessage(payload)
}

func (g AddrPayload) CommandName() CommandName {
	return AddrCommand
}

func (g *AddrPayload) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)

	addrCountEncoded, err := VarInt(len(g.AddressList)).encode()
	if err != nil {
		return nil, err
	}
	buffer.Write(addrCountEncoded)

	for _, a := range g.AddressList {
		err = binary.Write(buffer, binary.LittleEndian, a.Timestamp)
		if err != nil {
			return nil, err
		}
		netAddrEncoded, err := a.NetworkAddress.encode()
		if err != nil {
			return nil, err
		}
		buffer.Write(netAddrEncoded)
	}

	return buffer.Bytes(), nil
}

func decodeAddrPayload(r io.Reader) (*AddrPayload, error) {
	//answer, _ := io.ReadAll(r)
	//fmt.Println("answer", hex.EncodeToString(answer))
	addrCount, err := decodeVarInt(r)
	if err != nil {
		return nil, err
	}
	if addrCount > maxAddrCount {
		return nil, errors.New("exceeded max address count")
	}

	addressList := make([]Address, addrCount)
	for i := range addrCount {
		err = binary.Read(r, binary.LittleEndian, &addressList[i].Timestamp)
		if err != nil {
			return nil, err
		}
		netAddr, err := decodeNetworkAddress(r)
		if err != nil {
			return nil, err
		}
		addressList[i].NetworkAddress = *netAddr
	}

	return &AddrPayload{AddressList: addressList}, nil
}
