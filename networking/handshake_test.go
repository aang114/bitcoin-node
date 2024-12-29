package networking

import (
	"github.com/aang114/bitcoin-node/constants"
	"github.com/aang114/bitcoin-node/message"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net"
	"sync"
	"testing"
	"time"
)

func sendMsg(t *testing.T, conn net.Conn, msg *message.Message) {
	encodedMsg, err := msg.Encode()
	require.NoError(t, err)
	n, err := conn.Write(encodedMsg)
	require.NoError(t, err)
	require.Equal(t, len(encodedMsg), n)
}

func receiveMsg(t *testing.T, conn net.Conn) *message.Message {
	msg, err := message.DecodeMessage(conn)
	require.NoError(t, err)

	return msg
}

type HandshakeData struct {
	peerAddr                       net.TCPAddr
	tcpTimeout                     time.Duration
	peerVersionMsg                 *message.Message
	verackMsg                      *message.Message
	wtxidrelayMsg                  *message.Message
	peerVersionMsgWithVersion70016 *message.Message
}

func CreateHandshakeData(t *testing.T) *HandshakeData {
	h := HandshakeData{}

	h.peerAddr = net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5000}
	h.tcpTimeout = 10 * time.Second

	var err error
	h.peerVersionMsg, err = message.NewVersionMessage(
		70015,
		message.NodeNetwork,
		100,
		*message.NewNetworkAddress(message.NodeNetwork, net.ParseIP("0.0.0.0"), uint16(0)),
		*message.NewNetworkAddress(message.NodeNetwork, h.peerAddr.IP, uint16(h.peerAddr.Port)),
		200,
		"/Peer:0.0.1",
		300,
		false,
	)
	if err != nil {
		t.Fatal(err.Error())
	}

	h.verackMsg, err = message.NewVerackMessage()
	if err != nil {
		t.Fatal(err.Error())
	}

	h.wtxidrelayMsg, err = message.NewWtxidRelayMessage()
	if err != nil {
		t.Fatal(err.Error())
	}

	// version msg with version ≥ 70016
	h.peerVersionMsgWithVersion70016, err = message.NewVersionMessage(
		70016,
		message.NodeNetwork,
		100,
		*message.NewNetworkAddress(message.NodeNetwork, net.ParseIP("0.0.0.0"), uint16(0)),
		*message.NewNetworkAddress(message.NodeNetwork, h.peerAddr.IP, uint16(h.peerAddr.Port)),
		200,
		"/Peer:0.0.1",
		300,
		false,
	)
	if err != nil {
		t.Fatal(err.Error())
	}

	return &h
}

type HandshakeTestSuite struct {
	suite.Suite
	HandshakeData
}

func TestHandshakeTestSuite(t *testing.T) {
	suite.Run(t, &HandshakeTestSuite{})
}

func (s *HandshakeTestSuite) SetupSuite() {
	s.HandshakeData = *CreateHandshakeData(s.T())
}

func (s *HandshakeTestSuite) TestPerformHandshake_ShouldWork() {
	ln, err := net.Listen("tcp", s.peerAddr.String())
	if err != nil {
		s.FailNow(err.Error())
	}
	defer ln.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		conn, err := ln.Accept()
		s.NoError(err)
		defer conn.Close()

		// receive version msg
		msg := receiveMsg(s.T(), conn)
		s.Equal(message.VersionCommand, msg.Payload.CommandName())
		payload, ok := msg.Payload.(*message.VersionPayload)
		s.True(ok)
		s.Equal(constants.ProtocolVersion, payload.Version)
		s.Equal(constants.UserAgent, payload.UserAgent)

		// send version msg
		sendMsg(s.T(), conn, s.peerVersionMsg)

		// receive verack msg
		msg = receiveMsg(s.T(), conn)
		s.Equal(s.verackMsg, msg)

		// send verack msg
		sendMsg(s.T(), conn, s.verackMsg)
	}()

	// handshake should work
	conn, err := PerformHandshake(&s.peerAddr, s.tcpTimeout, message.NodeNetwork, message.NodeNetwork)
	s.NoError(err)
	defer conn.Close()
	s.Equal(s.peerAddr.String(), conn.RemoteAddr().String())

	wg.Wait()

}

func (s *HandshakeTestSuite) TestPerformHandshake_ShouldExchangeWtxidRelayWithVersion70016() {
	ln, err := net.Listen("tcp", s.peerAddr.String())
	if err != nil {
		s.FailNow(err.Error())
	}
	defer ln.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		// Accept connection to node
		conn, err := ln.Accept()
		s.NoError(err)
		defer conn.Close()

		// receive version msg
		msg := receiveMsg(s.T(), conn)
		s.Equal(message.VersionCommand, msg.Payload.CommandName())
		versionPayload, ok := msg.Payload.(*message.VersionPayload)
		s.True(ok)
		s.Equal(constants.ProtocolVersion, versionPayload.Version)
		s.Equal(constants.UserAgent, versionPayload.UserAgent)

		// send version msg
		sendMsg(s.T(), conn, s.peerVersionMsgWithVersion70016)

		// receive wtxidreelay msg
		msg = receiveMsg(s.T(), conn)
		s.Equal(s.wtxidrelayMsg, msg)

		// send wtxidrelay msg
		sendMsg(s.T(), conn, s.wtxidrelayMsg)

		// receive verack msg
		msg = receiveMsg(s.T(), conn)
		s.Equal(s.verackMsg, msg)

		// send verack msg
		sendMsg(s.T(), conn, s.verackMsg)
	}()

	// handshake should work
	conn, err := PerformHandshake(&s.peerAddr, s.tcpTimeout, message.NodeNetwork, message.NodeNetwork)
	s.NoError(err)
	defer conn.Close()
	s.Equal(s.peerAddr.String(), conn.RemoteAddr().String())

	wg.Wait()
}
