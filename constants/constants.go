package constants

import (
	"encoding/hex"
)

const (
	ProtocolVersion   int32  = 70015
	MainnetMagicValue        = uint32(0xD9B4BEF9)
	UserAgent         string = "/bitcoin-node-go:0.0.1/"
)

// https://bitcoinexplorer.org/block/000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f
var GenesisBlockHash, _ = hex.DecodeString("6fe28c0ab6f1b372c1a6a246ae63f74f931e8365e15a089c68d6190000000000")
