package main

import (
	"context"
	"github.com/aang114/bitcoin-node/constants"
	"github.com/aang114/bitcoin-node/message"
	"github.com/aang114/bitcoin-node/networking"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	// https://bitnodes.io/nodes/46.166.142.2:8333/
	remoteAddr := net.TCPAddr{IP: net.ParseIP("46.166.142.2"), Port: 8333}

	node := networking.NewNode(
		uint32(constants.ProtocolVersion),
		message.NodeNetwork,
		5,
		constants.BlocksFileDirectory,
		20*time.Second,
		10*time.Second,
		10*time.Second,
	)

	_, err := node.AddPeer(&remoteAddr, message.NodeNetwork)
	if err != nil {
		log.Fatalf("Adding Peer failed with error: %s", err)
	}

	go node.Start()

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer stop()

	select {
	case <-node.QuitCh:
		log.Println("Node has quit due to an error to an unresolvable error. Shutting down now...")
	case <-ctx.Done():
		log.Println("User sent a signal to quit the node. Shutting down now...")
		node.Quit()
		<-node.QuitCh
	}

	log.Println("Goodbye!")
}
