package main

import (
	"github.com/aang114/bitcoin-node/constants"
	"github.com/aang114/bitcoin-node/message"
	"github.com/aang114/bitcoin-node/networking"
	"log"
	"net"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	log.Println("Hello World")

	// https://bitnodes.io/nodes/46.166.142.2:8333/
	remoteAddr := net.TCPAddr{IP: net.ParseIP("46.166.142.2"), Port: 8333}

	node := networking.NewNode(
		uint32(constants.ProtocolVersion),
		message.NodeNetwork,
		5,
		20*time.Second,
		10*time.Second,
		10*time.Second,
	)

	_, err := node.AddPeer(&remoteAddr, message.NodeNetwork)
	if err != nil {
		log.Fatalf("Adding Peer failed with error: %s", err)
	}

	go node.Start()

	<-node.QuitCh
}
