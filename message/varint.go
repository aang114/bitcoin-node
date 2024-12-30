package message

import (
	"bytes"
	"encoding/binary"
	"io"
)

// https://en.bitcoin.it/wiki/Protocol_documentation#Variable_length_integer
type VarInt uint64

func (v VarInt) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)

	if v < 0xFD {
		binary.Write(buffer, binary.LittleEndian, uint8(v))
	} else if v <= 0xFFFF {
		_, err := buffer.Write([]byte{0xFD})
		if err != nil {
			return nil, err
		}
		err = binary.Write(buffer, binary.LittleEndian, uint16(v))
		if err != nil {
			return nil, err
		}
	} else if v <= 0xFFFF_FFFF {
		_, err := buffer.Write([]byte{0xFE})
		if err != nil {
			return nil, err
		}
		err = binary.Write(buffer, binary.LittleEndian, uint32(v))
		if err != nil {
			return nil, err
		}
	} else {
		_, err := buffer.Write([]byte{0xFF})
		if err != nil {
			return nil, err
		}
		err = binary.Write(buffer, binary.LittleEndian, v)
		if err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}

// https://en.bitcoin.it/wiki/Protocol_documentation#Variable_length_integer
func DecodeVarInt(r io.Reader) (VarInt, error) {
	buf := make([]byte, 1)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return 0, err
	}
	var number uint64
	switch buf[0] {
	case 0xFD:
		var n uint16
		err = binary.Read(r, binary.LittleEndian, &n)
		if err != nil {
			return 0, err
		}
		number = uint64(n)
	case 0xFE:
		var n uint32
		err = binary.Read(r, binary.LittleEndian, &n)
		if err != nil {
			return 0, err
		}
		number = uint64(n)
	case 0xFF:
		err = binary.Read(r, binary.LittleEndian, &number)
		if err != nil {
			return 0, err
		}
	default:
		number = uint64(buf[0])
	}

	return VarInt(number), nil
}
