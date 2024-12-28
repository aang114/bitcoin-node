package networking

import (
	"errors"
	"fmt"
	"github.com/aang114/bitcoin-node/message"
	"log"
	"net"
	"sync"
)

var ErrInvalidPayload = errors.New("invalid payload")

type TCPAddress struct {
	IpAddress [16]byte
	Port      uint16
}

func (t TCPAddress) String() string {
	return fmt.Sprintf("%s:%d", net.IP(t.IpAddress[:]), t.Port)
}

type Peer struct {
	mu                   sync.Mutex
	conn                 *net.TCPConn
	tcpAddress           TCPAddress
	HasQuit              bool
	onQuitting           func(*Peer)
	QuitCh               chan struct{}
	msgCh                chan *message.Message
	writeCh              chan []byte
	getAddrMsgResponseCh chan []message.Address
	invMsgCh             chan<- *InvPayloadWithSender
	blockMsgCh           chan<- *BlockPayloadWithSender
}

func NewPeer(conn *net.TCPConn, onQuitting func(*Peer), invMsgCh chan<- *InvPayloadWithSender, blockMsgCh chan<- *BlockPayloadWithSender) (*Peer, error) {
	addr, err := getRemoteAddr(conn)
	if err != nil {
		return nil, err
	}
	tcpAddress := TCPAddress{IpAddress: [16]byte(addr.IP.To16()), Port: uint16(addr.Port)}

	return &Peer{
		conn:       conn,
		tcpAddress: tcpAddress,
		HasQuit:    false,
		onQuitting: onQuitting,
		QuitCh:     make(chan struct{}),
		// TODO - Decide on the channel buffer length
		msgCh: make(chan *message.Message, 100),
		// TODO - Decide on the channel buffer length
		writeCh:              make(chan []byte, 100),
		getAddrMsgResponseCh: nil,
		invMsgCh:             invMsgCh,
		blockMsgCh:           blockMsgCh,
	}, nil
}

func (p *Peer) Start() {
	log.Printf("Starting Peer %s", p.conn.RemoteAddr())

	go p.readLoop()
	go p.msgChLoop()
	p.writeLoop()
}

func (p *Peer) Quit() {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Printf("Quitting Peer %s...", p.conn.RemoteAddr())

	if p.HasQuit {
		return
	}
	p.HasQuit = true

	if p.onQuitting != nil {
		p.onQuitting(p)
	}
	// closing the connection with close the readLoop()
	_ = p.conn.Close()

	close(p.QuitCh)
}

func (p *Peer) readLoop() {
	for {
		msg, err := message.DecodeMessage(p.conn)
		if err != nil {
			commandNameErr := &message.ErrUnknownCommandName{}
			if errors.As(err, &commandNameErr) {
				//log.Printf("[readLoop] Unknown Command Name: %s. Skipping...", commandNameErr.Command)
				continue
			} else {
				log.Printf("[readLoop] Quitting peer %s due to error: %s", p.conn.RemoteAddr(), err)
				p.Quit()
				return
			}
		}
		log.Printf("[readLoop] Read \"%s\" message from peer %s", msg.Header.Command, p.conn.RemoteAddr())
		p.msgCh <- msg
	}
}

func (p *Peer) msgChLoop() {
	for {
		select {
		case <-p.QuitCh:
			log.Printf("[msgChLoop] Peer %s's QuitCh was closed", p.conn.RemoteAddr())
			return
		case msg := <-p.msgCh:
			var err error
			switch msg.Header.Command {
			case message.PingCommand:
				err = p.handlePingMessage(msg)
			case message.AddrCommand:
				err = p.handleAddrMessage(msg)
			case message.InvCommand:
				err = p.handleInvMessage(msg)
			case message.BlockCommand:
				err = p.handleBlockMessage(msg)
			}
			if err != nil {
				//log.Printf("[msgChLoop] Quitting peer %s due to error: %s", p.conn.RemoteAddr(), err)
				p.Quit()
			} else {
				//log.Printf("[msgChLoop] Received Message \"%s\" from peer %s", msg.Header.Command, p.conn.RemoteAddr())
			}
		}
	}
}

func (p *Peer) writeLoop() {
	for {
		select {
		case <-p.QuitCh:
			//log.Printf("[writeLoop] Peer %s's QuitCh was closed", p.conn.RemoteAddr())
			return
		case bytes := <-p.writeCh:
			_, err := p.conn.Write(bytes)
			if err != nil {
				log.Printf("[writeLoop] Quitting peer %s due to error: %s", p.conn.RemoteAddr(), err)
			} else {
				//log.Printf("[writeLoop] Wrote %d-bytes message to peer %s", len(bytes), p.conn.RemoteAddr())
			}
		}
	}
}

func (p *Peer) handlePingMessage(msg *message.Message) error {
	pingPayload, ok := msg.Payload.(*message.PingPayload)
	if !ok {
		return ErrInvalidPayload
	}
	pongMsg, err := message.NewPongMessage(pingPayload.Nonce)
	if err != nil {
		return err
	}
	pongMsgEncoded, err := pongMsg.Encode()
	if err != nil {
		return err
	}
	p.write(pongMsgEncoded)

	return nil
}

func (p *Peer) handleAddrMessage(msg *message.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.getAddrMsgResponseCh == nil {
		return nil
	}

	addrPayload, ok := msg.Payload.(*message.AddrPayload)
	if !ok {
		return ErrInvalidPayload
	}

	// Each peer which wants to accept incoming connections creates an “addr” or “addrv2” message providing its connection information and then sends that message to its peers unsolicited (https://developer.bitcoin.org/reference/p2p_networking.html#addr)
	if len(addrPayload.AddressList) == 1 {
		if a := addrPayload.AddressList[0]; [16]byte(a.NetworkAddress.IpAddress.To16()) == p.tcpAddress.IpAddress && a.NetworkAddress.Port == p.tcpAddress.Port {
			return nil
		}
	}

	log.Printf("Solicited addr message from peer %s has %d addresses", p.conn.RemoteAddr(), len(addrPayload.AddressList))

	p.getAddrMsgResponseCh <- addrPayload.AddressList
	close(p.getAddrMsgResponseCh)
	p.getAddrMsgResponseCh = nil

	return nil
}

func (p *Peer) handleInvMessage(msg *message.Message) error {
	invPayload, ok := msg.Payload.(*message.InvPayload)
	if !ok {
		return ErrInvalidPayload
	}

	p.invMsgCh <- &InvPayloadWithSender{Sender: p, InvPayload: invPayload}

	return nil
}

func (p *Peer) handleBlockMessage(msg *message.Message) error {
	blockPayload, ok := msg.Payload.(*message.BlockPayload)
	if !ok {
		return ErrInvalidPayload
	}

	p.blockMsgCh <- &BlockPayloadWithSender{Sender: p, BlockPayload: blockPayload}

	return nil
}

func (p *Peer) write(bytes []byte) {
	p.writeCh <- bytes
}

func (p *Peer) sendGetAddrMsg() (<-chan []message.Address, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.getAddrMsgResponseCh = make(chan []message.Address)

	getAddrMsg, err := message.NewGetAddrMessage()
	if err != nil {
		return nil, err
	}
	getAddrMsgEncoded, err := getAddrMsg.Encode()
	p.write(getAddrMsgEncoded)

	log.Printf("╰┈➤ Sent getaddr message to peer %s", p.conn.RemoteAddr())

	return p.getAddrMsgResponseCh, nil
}

func (p *Peer) sendGetBlockDataMsg(blockInventories []message.Inventory) error {
	getDataMsg, err := message.NewGetDataMessage(blockInventories)
	if err != nil {
		return err
	}
	getDataMsgEncoded, err := getDataMsg.Encode()
	if err != nil {
		return err
	}
	p.write(getDataMsgEncoded)

	log.Printf("╰┈➤ Sent getdata Message to peer %s", p.conn.RemoteAddr())

	return nil
}

func (p *Peer) sendGetBlocksMsg(protocolVersion uint32, blockLocatorHashes []message.Hash256, stopHash message.Hash256) error {
	getBlocksMsg, err := message.NewGetBlocksMessage(protocolVersion, blockLocatorHashes, stopHash)
	if err != nil {
		return err
	}
	getBlocksMsgEncoded, err := getBlocksMsg.Encode()
	if err != nil {
		return err
	}
	p.write(getBlocksMsgEncoded)

	log.Printf("╰┈➤ Sent getblocks Message to peer %s", p.conn.RemoteAddr())

	return nil
}
