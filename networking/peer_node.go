package networking

import (
	"errors"
	"github.com/aang114/bitcoin-node/message"
	"log"
	"net"
	"sync"
)

type PeerNode struct {
	mu                sync.Mutex
	msgCh             chan message.Message
	conn              *net.TCPConn
	HasQuit           bool
	getAddrResponseCh chan []message.Address
	QuitCh            chan struct{}
}

func NewPeerNode(conn *net.TCPConn) *PeerNode {
	return &PeerNode{
		conn:  conn,
		msgCh: make(chan message.Message, 1),
	}
}

func (p *PeerNode) Start() {
	go p.readLoop()
	go p.msgChLoop()
}

// TODO - can this be a data race?
func (p *PeerNode) Quit() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.HasQuit {
		return
	}
	p.HasQuit = true

	p.conn.Close()
}

func (p *PeerNode) readLoop() {
	for {
		msg, err := message.DecodeMessage(p.conn)
		if err != nil {
			log.Printf("Read error from %s", err)
			continue
		}
		log.Printf("Read %s from %s", msg.Header.Command, p.conn.RemoteAddr())
		p.msgCh <- *msg
	}
}

func (p *PeerNode) msgChLoop() {
	for msg := range p.msgCh {
		log.Printf("Received  Message %s from %s", msg.Header.Command, p.conn.RemoteAddr())
		var err error
		switch msg.Header.Command {
		case message.AddrCommand:
			err = p.handleAddrMessage(msg)
		}
		if err != nil {
			log.Println("quitting peer due to error:", err)
			p.Quit()
		}
	}
}

func (p *PeerNode) SendGetAddrMsg() (<-chan []message.Address, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// TODO - Should I close existing channel, if it exists
	p.getAddrResponseCh = make(chan []message.Address)

	getAddrMsg, err := message.NewGetAddrMessage()
	if err != nil {
		return nil, err
	}
	getAddrMsgEncoded, err := getAddrMsg.Encode()
	_, err = p.conn.Write(getAddrMsgEncoded)
	if err != nil {
		return nil, err
	}

	log.Printf("sent getaddr message to peer node (%s)", p.conn.RemoteAddr())

	return p.getAddrResponseCh, nil
}

func (p *PeerNode) handleAddrMessage(msg message.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.getAddrResponseCh == nil {
		return nil
	}

	addrPayload, ok := msg.Payload.(*message.AddrPayload)
	if !ok {
		// TODO
		return errors.New("invalid payload")
	}
	remoteAddr, err := getRemoteAddr(p.conn)
	if err != nil {
		return err
	}

	// Each peer which wants to accept incoming connections creates an “addr” or “addrv2” message providing its connection information and then sends that message to its peers unsolicited (https://developer.bitcoin.org/reference/p2p_networking.html#addr)
	if len(addrPayload.AddressList) == 1 {
		if a := addrPayload.AddressList[0]; a.NetworkAddress.IpAddress.String() == remoteAddr.IP.String() && a.NetworkAddress.Port == uint16(remoteAddr.Port) {
			return nil
		}
	}
	log.Println("solicited addr message from %s has %d addresses", len(addrPayload.AddressList))

	p.getAddrResponseCh <- addrPayload.AddressList
	close(p.getAddrResponseCh)

	return nil
}
