package message

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type GetDataPayload struct {
	InventoryList []Inventory
}

func (p *GetDataPayload) CommandName() CommandName {
	return GetDataCommand
}

func newGetDataPayload(inventoryList []Inventory) *GetDataPayload {
	return &GetDataPayload{InventoryList: inventoryList}
}

func NewGetDataMessage(inventoryList []Inventory) (*Message, error) {
	payload := newGetDataPayload(inventoryList)
	return newMessage(payload)
}

func (p *GetDataPayload) Encode() ([]byte, error) {
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

func decodeGetDataPayload(r io.Reader) (*GetDataPayload, error) {
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

	return &GetDataPayload{InventoryList: inventoryList}, nil
}
