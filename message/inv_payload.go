package message

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const maxInvCount = 50_000

type Hash256 = [32]byte

type InventoryType uint32

const (
	Error                   InventoryType = 0
	MsgTx                   InventoryType = 1
	MsgBlock                InventoryType = 2
	MsgFilteredBlock        InventoryType = 3
	MsgCmpctBlock           InventoryType = 4
	MsgWitnessTx            InventoryType = 0x40000001
	MsgWitnessBlock         InventoryType = 0x40000002
	MsgFilteredWitnessBlock InventoryType = 0x40000003
)

type Inventory struct {
	Type InventoryType
	Hash [32]byte
}

type InvPayload struct {
	InventoryList []Inventory
}

func newInvPayload(inventoryList []Inventory) *InvPayload {
	return &InvPayload{InventoryList: inventoryList}
}

func NewInvMessage(inventoryList []Inventory) (*Message, error) {
	payload := newInvPayload(inventoryList)
	return newMessage(payload)
}

func (p *InvPayload) CommandName() CommandName {
	return InvCommand
}

func (p *InvPayload) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)

	countEncoded, err := VarInt(len(p.InventoryList)).encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(countEncoded)
	if err != nil {
		return nil, err
	}

	for _, i := range p.InventoryList {
		err = binary.Write(buffer, binary.LittleEndian, i.Type)
		if err != nil {
			return nil, err
		}
		_, err = buffer.Write(i.Hash[:])
		if err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}

func decodeInvPayload(r io.Reader) (*InvPayload, error) {
	count, err := decodeVarInt(r)
	if err != nil {
		return nil, err
	}
	if count > maxInvCount {
		return nil, errors.New("exceeded max inv count")
	}

	inventoryList := make([]Inventory, count)
	for i := range count {
		err = binary.Read(r, binary.LittleEndian, &inventoryList[i].Type)
		if err != nil {
			return nil, err
		}
		_, err = io.ReadFull(r, inventoryList[i].Hash[:])
		if err != nil {
			return nil, err
		}
	}

	return &InvPayload{InventoryList: inventoryList}, nil
}
