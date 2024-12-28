package message_test

import (
	"bytes"
	"encoding/hex"
	"github.com/aang114/bitcoin-node/message"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestMessage_Encode(t *testing.T) {
	t.Run("version message should encode", func(t *testing.T) {
		// Hexdump example of version message taken from https://en.bitcoin.it/wiki/Protocol_documentation#version
		expected, err := hex.DecodeString("F9BEB4D976657273696F6E000000000065000000030ECC5762EA0000010000000000000011B2D05000000000010000000000000000000000000000000000FFFF000000000000010000000000000000000000000000000000FFFF0000000000003B2EB35D8CE617650F2F5361746F7368693A302E372E322FC03E030000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		msg, err := message.NewVersionMessage(60002,
			message.NodeNetwork,
			1355854353,
			*message.NewNetworkAddress(message.NodeNetwork, net.ParseIP("0000:0000:0000:0000:0000:FFFF:0000:0000"), 0),
			*message.NewNetworkAddress(message.NodeNetwork, net.ParseIP("0000:0000:0000:0000:0000:FFFF:0000:0000"), 0),
			0x6517E68C5DB32E3B,
			"/Satoshi:0.7.2/",
			212672,
			false)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		encoded, err := msg.Encode()

		assert.NoError(t, err)
		assert.Equal(t, expected, encoded)
	})

	t.Run("verack message should encode", func(t *testing.T) {
		// Hexdump example of verack message taken from https://en.bitcoin.it/wiki/Protocol_documentation#verack
		expected, err := hex.DecodeString("F9BEB4D976657261636B000000000000000000005DF6E0E2")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		msg, err := message.NewVerackMessage()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		encoded, err := msg.Encode()

		assert.NoError(t, err)
		assert.Equal(t, expected, encoded)
	})

	t.Run("getaddr message should encode", func(t *testing.T) {
		// Equivalent to the hexdump example of verack message (https://en.bitcoin.it/wiki/Protocol_documentation#verack), apart from the command name
		expected, err := hex.DecodeString("F9BEB4D9676574616464720000000000000000005DF6E0E2")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		msg, err := message.NewGetAddrMessage()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		encoded, err := msg.Encode()

		assert.NoError(t, err)
		assert.Equal(t, expected, encoded)
	})

	t.Run("addr message should encode", func(t *testing.T) {
		// Hexdump example of addr message taken from https://en.bitcoin.it/wiki/Protocol_documentation#addr
		expected, err := hex.DecodeString("F9BEB4D96164647200000000000000001F000000ED52399B01E215104D010000000000000000000000000000000000FFFF0A000001208D")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		address := message.NewAddress(1292899810, *message.NewNetworkAddress(message.NodeNetwork, net.ParseIP("10.0.0.1"), 8333))
		msg, err := message.NewAddrMessage([]message.Address{*address})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		encoded, err := msg.Encode()

		assert.NoError(t, err)
		assert.Equal(t, expected, encoded)
	})

	t.Run("getblocks message should encode", func(t *testing.T) {
		// Hexdump example of getblocks message taken from https://developer.bitcoin.org/reference/p2p_networking.html#getblocks
		expected, err := hex.DecodeString("F9BEB4D9676574626C6F636B7300000065000000452A46487111010002D39F608A7775B537729884D4E6633BB2105E55A16A14D31B00000000000000005C3E6403D40837110A2E8AFB602B1C01714BDA7CE23BEA0A00000000000000000000000000000000000000000000000000000000000000000000000000000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		blockLocatorHash1, err := hex.DecodeString("D39F608A7775B537729884D4E6633BB2105E55A16A14D31B0000000000000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		blockLocatorHash2, err := hex.DecodeString("5C3E6403D40837110A2E8AFB602B1C01714BDA7CE23BEA0A0000000000000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		stopHash, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		msg, err := message.NewGetBlocksMessage(
			70001,
			[]message.Hash256{message.Hash256(blockLocatorHash1), message.Hash256(blockLocatorHash2)},
			message.Hash256(stopHash))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		encoded, err := msg.Encode()

		assert.NoError(t, err)
		assert.Equal(t, expected, encoded)
	})

	t.Run("inv message should encode", func(t *testing.T) {
		// Hexdump example of inv message taken from https://developer.bitcoin.org/reference/p2p_networking.html#inv
		expected, err := hex.DecodeString("F9BEB4D9696E76000000000000000000490000006467A0900201000000DE55FFD709AC1F5DC509A0925D0B1FC442CA034F224732E429081DA1B621F55A0100000091D36D997037E08018262978766F24B8A055AAF1D872E94AE85E9817B2C68DC7")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		txHash1, err := hex.DecodeString("DE55FFD709AC1F5DC509A0925D0B1FC442CA034F224732E429081DA1B621F55A")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		inventory1 := message.Inventory{Type: message.MsgTx, Hash: message.Hash256(txHash1)}
		txHash2, err := hex.DecodeString("91D36D997037E08018262978766F24B8A055AAF1D872E94AE85E9817B2C68DC7")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		inventory2 := message.Inventory{Type: message.MsgTx, Hash: message.Hash256(txHash2)}
		msg, err := message.NewInvMessage([]message.Inventory{inventory1, inventory2})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		encoded, err := msg.Encode()

		assert.NoError(t, err)
		assert.Equal(t, expected, encoded)
	})

	t.Run("getdata message should encode", func(t *testing.T) {
		// Equivalent to the hexdump example of inv message (https://en.bitcoin.it/wiki/Protocol_documentation#inv), apart from the command name
		expected, err := hex.DecodeString("F9BEB4D9676574646174610000000000490000006467A0900201000000DE55FFD709AC1F5DC509A0925D0B1FC442CA034F224732E429081DA1B621F55A0100000091D36D997037E08018262978766F24B8A055AAF1D872E94AE85E9817B2C68DC7")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		txHash1, err := hex.DecodeString("DE55FFD709AC1F5DC509A0925D0B1FC442CA034F224732E429081DA1B621F55A")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		inventory1 := message.Inventory{Type: message.MsgTx, Hash: message.Hash256(txHash1)}
		txHash2, err := hex.DecodeString("91D36D997037E08018262978766F24B8A055AAF1D872E94AE85E9817B2C68DC7")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		inventory2 := message.Inventory{Type: message.MsgTx, Hash: message.Hash256(txHash2)}
		msg, err := message.NewGetDataMessage([]message.Inventory{inventory1, inventory2})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		encoded, err := msg.Encode()

		assert.NoError(t, err)
		assert.Equal(t, expected, encoded)
	})

	t.Run("tx message should encode", func(t *testing.T) {
		// Hexdump example of tx message taken from https://en.bitcoin.it/wiki/Protocol_documentation#tx
		expected, err := hex.DecodeString("F9BEB4D974780000000000000000000002010000E293CDBE01000000016DBDDB085B1D8AF75184F0BC01FAD58D1266E9B63B50881990E4B40D6AEE3629000000008B483045022100F3581E1972AE8AC7C7367A7A253BC1135223ADB9A468BB3A59233F45BC578380022059AF01CA17D00E41837A1D58E97AA31BAE584EDEC28D35BD96923690913BAE9A0141049C02BFC97EF236CE6D8FE5D94013C721E915982ACD2B12B65D9B7D59E20A842005F8FC4E02532E873D37B96F09D6D4511ADA8F14042F46614A4C70C0F14BEFF5FFFFFFFF02404B4C00000000001976A9141AA0CD1CBEA6E7458A7ABAD512A9D9EA1AFB225E88AC80FAE9C7000000001976A9140EAB5BEA436A0484CFAB12485EFDA0B78B4ECC5288AC00000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		previousOutput, err := hex.DecodeString("6DBDDB085B1D8AF75184F0BC01FAD58D1266E9B63B50881990E4B40D6AEE362900000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		signatureScript, err := hex.DecodeString("483045022100F3581E1972AE8AC7C7367A7A253BC1135223ADB9A468BB3A59233F45BC578380022059AF01CA17D00E41837A1D58E97AA31BAE584EDEC28D35BD96923690913BAE9A0141049C02BFC97EF236CE6D8FE5D94013C721E915982ACD2B12B65D9B7D59E20A842005F8FC4E02532E873D37B96F09D6D4511ADA8F14042F46614A4C70C0F14BEFF5")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		txIn := message.NewTxIn(*message.NewOutPoint(message.Hash256(previousOutput), 0), signatureScript, 0xFFFFFFFF)
		pkScript1, err := hex.DecodeString("76A9141AA0CD1CBEA6E7458A7ABAD512A9D9EA1AFB225E88AC")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		txOut1 := message.NewTxOut(5000000, pkScript1)
		pkScript2, err := hex.DecodeString("76A9140EAB5BEA436A0484CFAB12485EFDA0B78B4ECC5288AC")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		txOut2 := message.NewTxOut(3354000000, pkScript2)
		msg, err := message.NewTxMessage(1, []message.TxIn{*txIn}, []message.TxOut{*txOut1, *txOut2}, []message.TxWitness{}, 0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		encoded, err := msg.Encode()

		assert.NoError(t, err)
		assert.Equal(t, expected, encoded)
	})

	t.Run("block message should encode", func(t *testing.T) {
		// Hexdump example of block message taken from https://developer.bitcoin.org/reference/block_chain.html#block-headers
		expected, err := hex.DecodeString("F9BEB4D9626C6F636B00000000000000510000009184952902000000B6FF0B1B1680A2862A30CA44D346D9E8910D334BEB48CA0C00000000000000009D10AA52EE949386CA9385695F04EDE270DDA20810DECD12BC9B048AAAB3147124D95A5430C31B18FE9F086400")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		prevBlock, err := hex.DecodeString("B6FF0B1B1680A2862A30CA44D346D9E8910D334BEB48CA0C0000000000000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		merkleRoot, err := hex.DecodeString("9D10AA52EE949386CA9385695F04EDE270DDA20810DECD12BC9B048AAAB31471")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		msg, err := message.NewBlockMessage(2, message.Hash256(prevBlock), message.Hash256(merkleRoot), 1415239972, 0x181bc330, 0x64089ffe, []message.TxPayload{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		encoded, err := msg.Encode()

		assert.NoError(t, err)
		assert.Equal(t, expected, encoded)
	})

}

func TestDecodeMessage(t *testing.T) {
	t.Run("version message should decode", func(t *testing.T) {
		expected, err := message.NewVersionMessage(60002,
			message.NodeNetwork,
			1355854353,
			*message.NewNetworkAddress(message.NodeNetwork, net.ParseIP("0000:0000:0000:0000:0000:FFFF:0000:0000"), 0),
			*message.NewNetworkAddress(message.NodeNetwork, net.ParseIP("0000:0000:0000:0000:0000:FFFF:0000:0000"), 0),
			0x6517E68C5DB32E3B,
			"/Satoshi:0.7.2/",
			212672,
			false)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Hexdump example of version message taken from https://en.bitcoin.it/wiki/Protocol_documentation#version
		encoded, err := hex.DecodeString("F9BEB4D976657273696F6E000000000065000000030ECC5762EA0000010000000000000011B2D05000000000010000000000000000000000000000000000FFFF000000000000010000000000000000000000000000000000FFFF0000000000003B2EB35D8CE617650F2F5361746F7368693A302E372E322FC03E030000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		decodedMsg, err := message.DecodeMessage(bytes.NewReader(encoded))

		assert.NoError(t, err)
		assert.Equal(t, expected, decodedMsg)
	})

	t.Run("verack message should decode", func(t *testing.T) {
		expected, err := message.NewVerackMessage()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Hexdump example of verack message taken from https://en.bitcoin.it/wiki/Protocol_documentation#verack
		encoded, err := hex.DecodeString("F9BEB4D976657261636B000000000000000000005DF6E0E2")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		decodedMsg, err := message.DecodeMessage(bytes.NewReader(encoded))

		assert.NoError(t, err)
		assert.Equal(t, expected, decodedMsg)
	})

	t.Run("addr message should decode", func(t *testing.T) {
		address := message.NewAddress(1292899810, *message.NewNetworkAddress(message.NodeNetwork, net.ParseIP("10.0.0.1"), 8333))
		expected, err := message.NewAddrMessage([]message.Address{*address})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Hexdump example of addr message taken from https://en.bitcoin.it/wiki/Protocol_documentation#addr
		encoded, err := hex.DecodeString("F9BEB4D96164647200000000000000001F000000ED52399B01E215104D010000000000000000000000000000000000FFFF0A000001208D")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		decodedMsg, err := message.DecodeMessage(bytes.NewReader(encoded))

		assert.NoError(t, err)
		assert.Equal(t, expected, decodedMsg)
	})

	t.Run("getaddr message should decode", func(t *testing.T) {
		expected, err := message.NewGetAddrMessage()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Equivalent to the hexdump example of verack message (https://en.bitcoin.it/wiki/Protocol_documentation#verack), apart from the command name
		encoded, err := hex.DecodeString("F9BEB4D9676574616464720000000000000000005DF6E0E2")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		decodedMsg, err := message.DecodeMessage(bytes.NewReader(encoded))

		assert.NoError(t, err)
		assert.Equal(t, expected, decodedMsg)
	})

	t.Run("getblocks message should decode", func(t *testing.T) {
		blockLocatorHash1, err := hex.DecodeString("D39F608A7775B537729884D4E6633BB2105E55A16A14D31B0000000000000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		blockLocatorHash2, err := hex.DecodeString("5C3E6403D40837110A2E8AFB602B1C01714BDA7CE23BEA0A0000000000000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		stopHash, err := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		expected, err := message.NewGetBlocksMessage(
			70001,
			[]message.Hash256{message.Hash256(blockLocatorHash1), message.Hash256(blockLocatorHash2)},
			message.Hash256(stopHash))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Hexdump example of getblocks message taken from https://developer.bitcoin.org/reference/p2p_networking.html#getblocks
		encoded, err := hex.DecodeString("F9BEB4D9676574626C6F636B7300000065000000452A46487111010002D39F608A7775B537729884D4E6633BB2105E55A16A14D31B00000000000000005C3E6403D40837110A2E8AFB602B1C01714BDA7CE23BEA0A00000000000000000000000000000000000000000000000000000000000000000000000000000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		decodedMsg, err := message.DecodeMessage(bytes.NewReader(encoded))

		assert.NoError(t, err)
		assert.Equal(t, expected, decodedMsg)
	})

	t.Run("inv message should decode", func(t *testing.T) {
		txHash1, err := hex.DecodeString("DE55FFD709AC1F5DC509A0925D0B1FC442CA034F224732E429081DA1B621F55A")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		inventory1 := message.Inventory{Type: message.MsgTx, Hash: message.Hash256(txHash1)}
		txHash2, err := hex.DecodeString("91D36D997037E08018262978766F24B8A055AAF1D872E94AE85E9817B2C68DC7")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		inventory2 := message.Inventory{Type: message.MsgTx, Hash: message.Hash256(txHash2)}
		expected, err := message.NewInvMessage([]message.Inventory{inventory1, inventory2})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Hexdump example of inv message taken from https://developer.bitcoin.org/reference/p2p_networking.html#inv
		encoded, err := hex.DecodeString("F9BEB4D9696E76000000000000000000490000006467A0900201000000DE55FFD709AC1F5DC509A0925D0B1FC442CA034F224732E429081DA1B621F55A0100000091D36D997037E08018262978766F24B8A055AAF1D872E94AE85E9817B2C68DC7")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		decodedMsg, err := message.DecodeMessage(bytes.NewReader(encoded))

		assert.NoError(t, err)
		assert.Equal(t, expected, decodedMsg)
	})

	t.Run("getdata message should decode", func(t *testing.T) {
		txHash1, err := hex.DecodeString("DE55FFD709AC1F5DC509A0925D0B1FC442CA034F224732E429081DA1B621F55A")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		inventory1 := message.Inventory{Type: message.MsgTx, Hash: message.Hash256(txHash1)}
		txHash2, err := hex.DecodeString("91D36D997037E08018262978766F24B8A055AAF1D872E94AE85E9817B2C68DC7")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		inventory2 := message.Inventory{Type: message.MsgTx, Hash: message.Hash256(txHash2)}
		expected, err := message.NewGetDataMessage([]message.Inventory{inventory1, inventory2})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Equivalent to the hexdump example of inv message (https://en.bitcoin.it/wiki/Protocol_documentation#inv), apart from the command name
		encoded, err := hex.DecodeString("F9BEB4D9676574646174610000000000490000006467A0900201000000DE55FFD709AC1F5DC509A0925D0B1FC442CA034F224732E429081DA1B621F55A0100000091D36D997037E08018262978766F24B8A055AAF1D872E94AE85E9817B2C68DC7")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		decodedMsg, err := message.DecodeMessage(bytes.NewReader(encoded))

		assert.NoError(t, err)
		assert.Equal(t, expected, decodedMsg)
	})

	t.Run("tx message should decode", func(t *testing.T) {
		previousOutput, err := hex.DecodeString("6DBDDB085B1D8AF75184F0BC01FAD58D1266E9B63B50881990E4B40D6AEE362900000000")
		if err != nil {
			t.Fatal(err)
		}
		signatureScript, err := hex.DecodeString("483045022100F3581E1972AE8AC7C7367A7A253BC1135223ADB9A468BB3A59233F45BC578380022059AF01CA17D00E41837A1D58E97AA31BAE584EDEC28D35BD96923690913BAE9A0141049C02BFC97EF236CE6D8FE5D94013C721E915982ACD2B12B65D9B7D59E20A842005F8FC4E02532E873D37B96F09D6D4511ADA8F14042F46614A4C70C0F14BEFF5")
		if err != nil {
			t.Fatal(err)
		}
		txIn := message.NewTxIn(*message.NewOutPoint(message.Hash256(previousOutput), 0), signatureScript, 0xFFFFFFFF)
		pkScript1, err := hex.DecodeString("76A9141AA0CD1CBEA6E7458A7ABAD512A9D9EA1AFB225E88AC")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		txOut1 := message.NewTxOut(5000000, pkScript1)
		pkScript2, err := hex.DecodeString("76A9140EAB5BEA436A0484CFAB12485EFDA0B78B4ECC5288AC")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		txOut2 := message.NewTxOut(3354000000, pkScript2)
		expected, err := message.NewTxMessage(1, []message.TxIn{*txIn}, []message.TxOut{*txOut1, *txOut2}, []message.TxWitness{}, 0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Hexdump example of tx message taken from https://en.bitcoin.it/wiki/Protocol_documentation#tx
		encodedMsg, err := hex.DecodeString("F9BEB4D974780000000000000000000002010000E293CDBE01000000016DBDDB085B1D8AF75184F0BC01FAD58D1266E9B63B50881990E4B40D6AEE3629000000008B483045022100F3581E1972AE8AC7C7367A7A253BC1135223ADB9A468BB3A59233F45BC578380022059AF01CA17D00E41837A1D58E97AA31BAE584EDEC28D35BD96923690913BAE9A0141049C02BFC97EF236CE6D8FE5D94013C721E915982ACD2B12B65D9B7D59E20A842005F8FC4E02532E873D37B96F09D6D4511ADA8F14042F46614A4C70C0F14BEFF5FFFFFFFF02404B4C00000000001976A9141AA0CD1CBEA6E7458A7ABAD512A9D9EA1AFB225E88AC80FAE9C7000000001976A9140EAB5BEA436A0484CFAB12485EFDA0B78B4ECC5288AC00000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		decodedMsg, err := message.DecodeMessage(bytes.NewReader(encodedMsg))

		assert.NoError(t, err)
		assert.Equal(t, expected, decodedMsg)
	})

	t.Run("block message should decode", func(t *testing.T) {
		prevBlock, err := hex.DecodeString("B6FF0B1B1680A2862A30CA44D346D9E8910D334BEB48CA0C0000000000000000")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		merkleRoot, err := hex.DecodeString("9D10AA52EE949386CA9385695F04EDE270DDA20810DECD12BC9B048AAAB31471")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		expected, err := message.NewBlockMessage(2, message.Hash256(prevBlock), message.Hash256(merkleRoot), 1415239972, 0x181bc330, 0x64089ffe, []message.TxPayload{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Hexdump example of block message taken from https://developer.bitcoin.org/reference/block_chain.html#block-headers
		encoded, err := hex.DecodeString("F9BEB4D9626C6F636B00000000000000510000009184952902000000B6FF0B1B1680A2862A30CA44D346D9E8910D334BEB48CA0C00000000000000009D10AA52EE949386CA9385695F04EDE270DDA20810DECD12BC9B048AAAB3147124D95A5430C31B18FE9F086400")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		decodedMsg, err := message.DecodeMessage(bytes.NewReader(encoded))

		assert.NoError(t, err)
		assert.Equal(t, expected, decodedMsg)
	})
}
