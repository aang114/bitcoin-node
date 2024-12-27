package message

import (
	"bytes"
	"encoding/binary"
	"io"
)

// Return an inv packet containing the list of blocks starting right after the last known hash in the block locator object, up to hash_stop or 500 blocks, whichever comes first. (https://en.bitcoin.it/wiki/Protocol_documentation#getblocks)
type GetBlocksPayload struct {
	// The protocol version number; the same as sent in the “version” message.
	Version uint32
	// Hashes should be provided in reverse order of block height, so highest-height hashes are listed first and lowest-height hashes are listed last.
	BlockLocatorHashes []Hash256
	// Hash of the last desired block; set to zero to get as many blocks as possible (500)
	HashStop Hash256
}

func (p *GetBlocksPayload) CommandName() CommandName {
	return GetBlocksCommand
}

func (p *GetBlocksPayload) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, p.Version)
	if err != nil {
		return nil, err
	}
	blockLocatorHashesCountEncoded, err := VarInt(len(p.BlockLocatorHashes)).encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(blockLocatorHashesCountEncoded)
	if err != nil {
		return nil, err
	}
	for _, blockHash := range p.BlockLocatorHashes {
		_, err = buffer.Write(blockHash[:])
		if err != nil {
			return nil, err
		}
	}
	_, err = buffer.Write(p.HashStop[:])
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeGetBlocksPayload(r io.Reader) (*GetBlocksPayload, error) {
	p := GetBlocksPayload{}

	err := binary.Read(r, binary.LittleEndian, &p.Version)
	if err != nil {
		return nil, err
	}
	blockLocatorHashesCount, err := decodeVarInt(r)
	if err != nil {
		return nil, err
	}
	p.BlockLocatorHashes = make([]Hash256, blockLocatorHashesCount)
	for i := range p.BlockLocatorHashes {
		_, err = io.ReadFull(r, p.BlockLocatorHashes[i][:])
		if err != nil {
			return nil, err
		}
	}
	_, err = io.ReadFull(r, p.HashStop[:])
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func newGetBlocksPayload(version uint32, blockLocatorHashes []Hash256, hashStop Hash256) *GetBlocksPayload {
	return &GetBlocksPayload{
		Version:            version,
		BlockLocatorHashes: blockLocatorHashes,
		HashStop:           hashStop,
	}
}

func NewGetBlocksMessage(version uint32, blockLocatorHashes []Hash256, hashStop Hash256) (*Message, error) {
	payload := newGetBlocksPayload(version, blockLocatorHashes, hashStop)
	return newMessage(payload)
}
