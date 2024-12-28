package message

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"io"
)

// https://en.bitcoin.it/wiki/Protocol_documentation#block
type BlockPayload struct {
	// Block version information (note, this is signed)
	Version int32
	// The hash value of the previous block this particular block references
	PrevBlock Hash256
	// The reference to a Merkle tree collection which is a hash of all transactions related to this block
	MerkleRoot Hash256
	// A Unix timestamp recording when this block was created (Currently limited to dates before the year 2106!)
	Timestamp uint32
	// The calculated difficulty target being used for this block
	Bits uint32
	// The nonce used to generate this block… to allow variations of the header and compute different hashes
	Nonce uint32
	// Block transactions, in format of "tx" command
	Transactions []TxPayload
}

func newBlockPayload(version int32, prevBlock Hash256, merkleRoot Hash256, timestamp uint32, bits uint32, nonce uint32, transactions []TxPayload) *BlockPayload {
	return &BlockPayload{
		Version:      version,
		PrevBlock:    prevBlock,
		MerkleRoot:   merkleRoot,
		Timestamp:    timestamp,
		Bits:         bits,
		Nonce:        nonce,
		Transactions: transactions,
	}
}

func NewBlockMessage(version int32, prevBlock Hash256, merkleRoot Hash256, timestamp uint32, bits uint32, nonce uint32, transactions []TxPayload) (*Message, error) {
	payload := newBlockPayload(version, prevBlock, merkleRoot, timestamp, bits, nonce, transactions)
	return newMessage(payload)
}

func (b *BlockPayload) CommandName() CommandName {
	return BlockCommand
}

func (b *BlockPayload) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, b.Version)
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(b.PrevBlock[:])
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(b.MerkleRoot[:])
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, b.Timestamp)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, b.Bits)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, b.Nonce)
	if err != nil {
		return nil, err
	}
	transactionsCount := VarInt(len(b.Transactions))
	encodedCount, err := transactionsCount.encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(encodedCount)
	if err != nil {
		return nil, err
	}
	for _, tx := range b.Transactions {
		txEncoded, err := tx.Encode()
		if err != nil {
			return nil, err
		}
		_, err = buffer.Write(txEncoded)
		if err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}

func decodeBlockPayload(r io.Reader) (*BlockPayload, error) {
	b := BlockPayload{}
	err := binary.Read(r, binary.LittleEndian, &b.Version)
	if err != nil {
		return nil, err
	}
	_, err = io.ReadFull(r, b.PrevBlock[:])
	if err != nil {
		return nil, err
	}
	_, err = io.ReadFull(r, b.MerkleRoot[:])
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &b.Timestamp)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &b.Bits)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &b.Nonce)
	if err != nil {
		return nil, err
	}
	transactionsCount, err := decodeVarInt(r)
	if err != nil {
		return nil, err
	}
	b.Transactions = make([]TxPayload, transactionsCount)
	for i := range transactionsCount {
		tx, err := decodeTxPayload(r)
		if err != nil {
			return nil, err
		}
		b.Transactions[i] = *tx
	}

	return &b, nil
}

// The SHA256 hash that identifies each block (and which must have a run of 0 bits) is calculated from the first 6 fields of this structure (version, prev_block, merkle_root, timestamp, bits, nonce, and standard SHA256 padding, making two 64-byte chunks in all) and not from the complete block (https://en.bitcoin.it/wiki/Protocol_documentation#block)
func (b *BlockPayload) GetBlockHash() (Hash256, error) {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, b.Version)
	if err != nil {
		return Hash256{}, err
	}
	_, err = buffer.Write(b.PrevBlock[:])
	if err != nil {
		return Hash256{}, err
	}
	_, err = buffer.Write(b.MerkleRoot[:])
	if err != nil {
		return Hash256{}, err
	}
	err = binary.Write(buffer, binary.LittleEndian, b.Timestamp)
	if err != nil {
		return Hash256{}, err
	}
	err = binary.Write(buffer, binary.LittleEndian, b.Bits)
	if err != nil {
		return Hash256{}, err
	}
	err = binary.Write(buffer, binary.LittleEndian, b.Nonce)
	if err != nil {
		return Hash256{}, err
	}

	hash := sha256.Sum256(buffer.Bytes())
	hash = sha256.Sum256(hash[:])

	return hash, nil
}
