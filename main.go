package main

import (
	"fmt"
	"github.com/aang114/bitcoin-node/message"
	"github.com/aang114/bitcoin-node/networking"
	"log"
	"net"
)

func main() {
	fmt.Println("Hello World")

	// https://bitnodes.io/nodes/73.65.210.23-8333/
	remoteAddr := net.TCPAddr{IP: net.ParseIP("73.65.210.23"), Port: 8333}

	err := networking.PerformHandshake(remoteAddr, message.NodeNetwork, message.NodeNetwork)
	if err != nil {
		log.Fatalln(err)
	}
}
