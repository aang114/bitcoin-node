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

}
