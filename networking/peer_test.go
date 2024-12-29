package networking

import (
	"bytes"
	"encoding/hex"
	"github.com/aang114/bitcoin-node/constants"
	"github.com/aang114/bitcoin-node/message"
	"github.com/stretchr/testify/suite"
	"log"
	"net"
	"sync"
	"testing"
)

type PeerTestSuite struct {
	suite.Suite
	HandshakeData
	nodeConn   net.Conn
	peerConn   net.Conn
	peer       *Peer
	invMsgCh   chan *InvPayloadWithSender
	blockMsgCh chan *BlockPayloadWithSender
	pingMsg    *message.Message
	invMsg     *message.Message
	blockMsg   *message.Message
	addrMsg    *message.Message
}

func TestPeerTestSuite(t *testing.T) {
	suite.Run(t, &PeerTestSuite{})
}

func (s *PeerTestSuite) SetupSuite() {
	s.HandshakeData = *CreateHandshakeData(s.T())

	var err error
	s.pingMsg, err = message.NewPingMessage(100)
	if err != nil {
		s.FailNow(err.Error())
	}

	// Hexdump example of inv message taken from https://developer.bitcoin.org/reference/p2p_networking.html#inv
	encodedInvMsg, err := hex.DecodeString("F9BEB4D9696E76000000000000000000490000006467A0900201000000DE55FFD709AC1F5DC509A0925D0B1FC442CA034F224732E429081DA1B621F55A0100000091D36D997037E08018262978766F24B8A055AAF1D872E94AE85E9817B2C68DC7")
	if err != nil {
		s.FailNow(err.Error())
	}
	s.invMsg, err = message.DecodeMessage(bytes.NewReader(encodedInvMsg))

	// Hexdump example of block message taken from https://developer.bitcoin.org/reference/block_chain.html#block-headers
	encodedBlockMsg, err := hex.DecodeString("F9BEB4D9626C6F636B00000000000000510000009184952902000000B6FF0B1B1680A2862A30CA44D346D9E8910D334BEB48CA0C00000000000000009D10AA52EE949386CA9385695F04EDE270DDA20810DECD12BC9B048AAAB3147124D95A5430C31B18FE9F086400")
	if err != nil {
		s.FailNow(err.Error())
	}
	s.blockMsg, err = message.DecodeMessage(bytes.NewReader(encodedBlockMsg))
	if err != nil {
		s.FailNow(err.Error())
	}

	// Hexdump example of addr message taken from https://en.bitcoin.it/wiki/Protocol_documentation#addr
	encodedAddrMsg, err := hex.DecodeString("F9BEB4D96164647200000000000000001F000000ED52399B01E215104D010000000000000000000000000000000000FFFF0A000001208D")
	s.addrMsg, err = message.DecodeMessage(bytes.NewReader(encodedAddrMsg))
	if err != nil {
		s.FailNow(err.Error())
	}
}

func performHandshake(s *PeerTestSuite) {
	var err error
	ln, err := net.Listen("tcp", s.peerAddr.String())
	defer ln.Close()
	if err != nil {
		s.FailNow(err.Error())
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		var err error
		s.peerConn, err = ln.Accept()
		s.NoError(err)

		// receive version msg
		msg := receiveMsg(s.T(), s.peerConn)
		s.Equal(message.VersionCommand, msg.Payload.CommandName())
		payload, ok := msg.Payload.(*message.VersionPayload)
		s.True(ok)
		s.Equal(constants.ProtocolVersion, payload.Version)
		s.Equal(constants.UserAgent, payload.UserAgent)

		// send version msg
		sendMsg(s.T(), s.peerConn, s.peerVersionMsg)

		// receive verack msg
		msg = receiveMsg(s.T(), s.peerConn)
		s.Equal(s.verackMsg, msg)

		// send verack msg
		sendMsg(s.T(), s.peerConn, s.verackMsg)
	}()

	s.nodeConn, err = PerformHandshake(&s.peerAddr, s.tcpTimeout, message.NodeNetwork, message.NodeNetwork)
	if err != nil {
		s.FailNow(err.Error())
	}
	if s.peerAddr.String() != s.nodeConn.RemoteAddr().String() {
		s.FailNow("unexpected peer address")
	}

	wg.Wait()
}

func setupPeer(s *PeerTestSuite, conn net.Conn) {
	s.invMsgCh = make(chan *InvPayloadWithSender, 100)
	s.blockMsgCh = make(chan *BlockPayloadWithSender, 100)
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		s.FailNow("peer conn is not tcp connection")
	}
	var err error
	s.peer, err = NewPeer(
		tcpConn,
		nil,
		s.invMsgCh,
		s.blockMsgCh,
	)
	if err != nil {
		s.FailNow(err.Error())
	}
}

func (s *PeerTestSuite) SetupTest() {
	performHandshake(s)
	setupPeer(s, s.nodeConn)
}

func (s *PeerTestSuite) TearDownTest() {
	log.Println("Tearing down test now...")
	s.peer.Quit()
	<-s.peer.QuitCh
	//s.nodeConn.Close()
	s.peerConn.Close()
}

func (s *PeerTestSuite) TestPeer_PingPongWorks() {
	go s.peer.Start()

	sendMsg(s.T(), s.peerConn, s.pingMsg)
	msg := receiveMsg(s.T(), s.peerConn)
	s.Equal(message.PongCommand, msg.Payload.CommandName())

	pingPayload, ok := s.pingMsg.Payload.(*message.PingPayload)
	s.True(ok)
	pongPayload, ok := msg.Payload.(*message.PongPayload)
	s.True(ok)
	s.Equal(pingPayload.Nonce, pongPayload.Nonce)
}

func (s *PeerTestSuite) TestPeer_InvMsgChWorks() {
	go s.peer.Start()

	sendMsg(s.T(), s.peerConn, s.invMsg)

	invMsgWithSender := <-s.invMsgCh

	s.Equal(s.peer, invMsgWithSender.Sender)
	s.Equal(s.invMsg.Payload, invMsgWithSender.InvPayload)
}

func (s *PeerTestSuite) TestPeer_BlockMsgChWorks() {
	go s.peer.Start()

	sendMsg(s.T(), s.peerConn, s.blockMsg)

	blockMsgWithSender := <-s.blockMsgCh

	s.Equal(s.peer, blockMsgWithSender.Sender)
	s.Equal(s.blockMsg.Payload, blockMsgWithSender.BlockPayload)
}

func (s *PeerTestSuite) TestPeer_GetAddrMsgResponseChWorks() {
	go s.peer.Start()

	getAddrMsgResponseCh, err := s.peer.sendGetAddrMsg()
	s.NoError(err)

	sendMsg(s.T(), s.peerConn, s.addrMsg)

	addresses := <-getAddrMsgResponseCh

	addrPayload, ok := s.addrMsg.Payload.(*message.AddrPayload)
	s.True(ok)
	s.Equal(addrPayload.AddressList, addresses)
}

func (s *PeerTestSuite) TestPeer_Quit() {
	go s.peer.Start()

	s.peerConn.Close()

	<-s.peer.QuitCh
	s.True(s.peer.HasQuit)
}
