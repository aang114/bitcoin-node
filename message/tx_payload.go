package message

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// https://github.com/bitcoin/bitcoin/blob/3f826598a42dcc707b58224e94c394e30a42ceee/src/script/script.h#L31-L32
const maxScriptSize = 10000

type OutPoint struct {
	// The hash of the referenced transaction.
	Hash Hash256
	// The index of the specific output in the transaction. The first output is 0, etc.
	Index uint32
}

func NewOutPoint(hash Hash256, index uint32) *OutPoint {
	return &OutPoint{
		Hash:  hash,
		Index: index,
	}
}

type TxIn struct {
	// The previous output transaction reference, as an OutPoint structure
	PreviousOutput OutPoint
	// Computational Script for confirming transaction authorization
	SignatureScript []byte
	// Transaction version as defined by the sender. Intended for "replacement" of transactions when information is updated before inclusion into a block.
	Sequence uint32
}

func NewTxIn(previousOutput OutPoint, signatureScript []byte, sequence uint32) *TxIn {
	return &TxIn{
		PreviousOutput:  previousOutput,
		SignatureScript: signatureScript,
		Sequence:        sequence,
	}
}

type TxOut struct {
	// Transaction Value
	Value int64
	// Usually contains the public key as a Bitcoin script setting up conditions to claim this output.
	PkScript []byte
}

func NewTxOut(value int64, pkScript []byte) *TxOut {
	return &TxOut{
		Value:    value,
		PkScript: pkScript,
	}
}

type ComponentData []byte

// The TxWitness structure consists of a var_int count of witness data components, followed by (for each witness data component) a var_int length of the component and the raw component data itself. (https://en.bitcoin.it/wiki/Protocol_documentation#tx)
type TxWitness struct {
	ComponentDataList []ComponentData
}

func NewTxWitness(componentDataList []ComponentData) *TxWitness {
	return &TxWitness{
		ComponentDataList: componentDataList,
	}
}

// https://en.bitcoin.it/wiki/Protocol_documentation#tx
type TxPayload struct {
	// Transaction data format version
	Version uint32
	// A list of 1 or more transaction inputs or sources for coins
	TransactionInputs []TxIn
	// A list of 1 or more transaction outputs or destinations for coins
	TransactionOutputs []TxOut
	// A list of witnesses, one for each input
	TransactionWitnesses []TxWitness
	// The block number or timestamp at which this transaction is unlocked
	LockTime uint32
}

func newTxPayload(version uint32, txInputs []TxIn, txOutputs []TxOut, txWitnesses []TxWitness, lockTime uint32) *TxPayload {
	return &TxPayload{
		Version:              version,
		TransactionInputs:    txInputs,
		TransactionOutputs:   txOutputs,
		TransactionWitnesses: txWitnesses,
		LockTime:             lockTime,
	}
}

func NewTxMessage(version uint32, txInputs []TxIn, txOutputs []TxOut, txWitnesses []TxWitness, lockTime uint32) (*Message, error) {
	payload := newTxPayload(version, txInputs, txOutputs, txWitnesses, lockTime)
	return newMessage(payload)
}

func (t *TxPayload) CommandName() CommandName {
	return TxCommand
}

func (t *TxPayload) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, t.Version)
	if err != nil {
		return nil, err
	}
	if len(t.TransactionWitnesses) > 0 {
		// If present, flag is always 0001, and indicates the presence of witness data
		flag := []byte{0x00, 0x01}
		_, err = buffer.Write(flag)
		if err != nil {
			return nil, err
		}
	}
	txInputsCount := VarInt(len(t.TransactionInputs))
	encodedCount, err := txInputsCount.Encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(encodedCount)
	if err != nil {
		return nil, err
	}
	for _, txIn := range t.TransactionInputs {
		encodedTxIn, err := txIn.Encode()
		if err != nil {
			return nil, err
		}
		_, err = buffer.Write(encodedTxIn)
		if err != nil {
			return nil, err
		}
	}
	txOutputsCount := VarInt(len(t.TransactionOutputs))
	encodedCount, err = txOutputsCount.Encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(encodedCount)
	if err != nil {
		return nil, err
	}
	for _, txOut := range t.TransactionOutputs {
		encodedTxOut, err := txOut.Encode()
		if err != nil {
			return nil, err
		}
		_, err = buffer.Write(encodedTxOut)
		if err != nil {
			return nil, err
		}
	}
	if len(t.TransactionWitnesses) > 0 {
		txWitnessesCount := VarInt(len(t.TransactionWitnesses))
		encodedCount, err = txWitnessesCount.Encode()
		if err != nil {
			return nil, err
		}
		_, err = buffer.Write(encodedCount)
		if err != nil {
			return nil, err
		}
		for _, txWitness := range t.TransactionWitnesses {
			encodedTxWitness, err := txWitness.Encode()
			if err != nil {
				return nil, err
			}
			_, err = buffer.Write(encodedTxWitness)
			if err != nil {
				return nil, err
			}
		}
	}
	err = binary.Write(buffer, binary.LittleEndian, t.LockTime)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeTxPayload(reader io.Reader) (*TxPayload, error) {
	r := bufio.NewReader(reader)

	t := TxPayload{}

	err := binary.Read(r, binary.LittleEndian, &t.Version)
	if err != nil {
		return nil, err
	}
	buf, err := r.Peek(2)
	if err != nil {
		return nil, err
	}
	flag := false
	if bytes.Equal(buf, []byte{0x00, 0x01}) {
		flag = true
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return nil, err
		}
	}
	txInputCount, err := DecodeVarInt(r)
	if err != nil {
		return nil, err
	}
	t.TransactionInputs = make([]TxIn, txInputCount)
	for i := range txInputCount {
		txIn, err := decodeTxIn(r)
		if err != nil {
			return nil, err
		}
		t.TransactionInputs[i] = *txIn
	}
	txOutputCount, err := DecodeVarInt(r)
	if err != nil {
		return nil, err
	}
	t.TransactionOutputs = make([]TxOut, txOutputCount)
	for i := range txOutputCount {
		txOut, err := decodeTxOut(r)
		if err != nil {
			return nil, err
		}
		t.TransactionOutputs[i] = *txOut
	}
	if flag {
		txWitnessCount, err := DecodeVarInt(r)
		if err != nil {
			return nil, err
		}
		t.TransactionWitnesses = make([]TxWitness, txWitnessCount)
		for i := range txWitnessCount {
			txWitness, err := decodeTxWitness(r)
			if err != nil {
				return nil, err
			}
			t.TransactionWitnesses[i] = *txWitness
		}
	} else {
		t.TransactionWitnesses = make([]TxWitness, 0)
	}
	err = binary.Read(r, binary.LittleEndian, &t.LockTime)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *OutPoint) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	_, err := buffer.Write(t.Hash[:])
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, t.Index)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeOutPoint(r io.Reader) (*OutPoint, error) {
	o := OutPoint{}
	_, err := io.ReadFull(r, o.Hash[:])
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &o.Index)
	if err != nil {
		return nil, err
	}

	return &o, nil
}

func (t *TxIn) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	previousOutputEncoded, err := t.PreviousOutput.Encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(previousOutputEncoded)
	if err != nil {
		return nil, err
	}
	scriptLengthEncoded, err := VarInt(len(t.SignatureScript)).Encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(scriptLengthEncoded)
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(t.SignatureScript)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.LittleEndian, t.Sequence)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeTxIn(r io.Reader) (*TxIn, error) {
	t := TxIn{}
	previousOutput, err := decodeOutPoint(r)
	if err != nil {
		return nil, err
	}
	t.PreviousOutput = *previousOutput
	scriptLength, err := DecodeVarInt(r)
	if err != nil {
		return nil, err
	}
	//log.Printf("scriptLength is %d", scriptLength)
	if scriptLength > maxScriptSize {
		return nil, errors.New(fmt.Sprintf("signatureScript (length: %d) exceeded max length", scriptLength))
	}
	t.SignatureScript = make([]byte, scriptLength)
	_, err = io.ReadFull(r, t.SignatureScript)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, binary.LittleEndian, &t.Sequence)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *TxOut) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, t.Value)
	if err != nil {
		return nil, err
	}
	pkScriptLengthEncoded, err := VarInt(len(t.PkScript)).Encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(pkScriptLengthEncoded)
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(t.PkScript)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decodeTxOut(r io.Reader) (*TxOut, error) {
	t := TxOut{}
	err := binary.Read(r, binary.LittleEndian, &t.Value)
	if err != nil {
		return nil, err
	}
	pkScriptLength, err := DecodeVarInt(r)
	if err != nil {
		return nil, err
	}
	//log.Printf("pkScriptLength is %d", pkScriptLength)
	if pkScriptLength > maxScriptSize {
		return nil, errors.New(fmt.Sprintf("pkScript (length %d) exceeded max length", pkScriptLength))
	}
	t.PkScript = make([]byte, pkScriptLength)
	_, err = io.ReadFull(r, t.PkScript)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *TxWitness) Encode() ([]byte, error) {
	buffer := new(bytes.Buffer)

	componentsCountEncoded, err := VarInt(len(t.ComponentDataList)).Encode()
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(componentsCountEncoded)
	if err != nil {
		return nil, err
	}
	for _, componentData := range t.ComponentDataList {
		componentDataLengthEncoded, err := VarInt(len(componentData)).Encode()
		if err != nil {
			return nil, err
		}
		_, err = buffer.Write(componentDataLengthEncoded)
		if err != nil {
			return nil, err
		}
		_, err = buffer.Write(componentData)
		if err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}

func decodeTxWitness(r io.Reader) (*TxWitness, error) {
	t := TxWitness{}
	componentsCount, err := DecodeVarInt(r)
	if err != nil {
		return nil, err
	}
	t.ComponentDataList = make([]ComponentData, componentsCount)
	for i := range componentsCount {
		componentDataLength, err := DecodeVarInt(r)
		if err != nil {
			return nil, err
		}
		t.ComponentDataList[i] = make(ComponentData, componentDataLength)
		_, err = io.ReadFull(r, t.ComponentDataList[i])
		if err != nil {
			return nil, err
		}
	}

	return &t, nil
}
