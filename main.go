package main

import (
	"github.com/aang114/bitcoin-node/message"
	"github.com/aang114/bitcoin-node/networking"
	"log"
	"net"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	log.Println("Hello World")

	// https://bitnodes.io/nodes/73.65.210.23-8333/
	remoteAddr := net.TCPAddr{IP: net.ParseIP("73.65.210.23"), Port: 8333}

	conn, err := networking.PerformHandshake(remoteAddr, message.NodeNetwork, message.NodeNetwork)
	if err != nil {
		log.Fatalln(err)
	}

	peer := networking.NewPeerNode(conn)
	go peer.Start()

	addressListCh, err := peer.SendGetAddrMsg()
	if err != nil {
		log.Fatalln(err)
	}

	addressList, ok := <-addressListCh
	if !ok {
		log.Fatalln("channel closed before sending address list")
	}
	log.Println("addressList:", addressList)

	for _, address := range addressList {
		tcpAddr := net.TCPAddr{IP: address.NetworkAddress.IpAddress[:], Port: int(address.NetworkAddress.Port)}
		log.Println("address:", tcpAddr)
	}

	<-peer.QuitCh

}
