package networking

import (
	"github.com/aang114/bitcoin-node/constants"
	"github.com/aang114/bitcoin-node/message"
	"github.com/stretchr/testify/suite"
	"net"
	"sync"
	"testing"
	"time"
)

type NodeTestSuite struct {
	suite.Suite
	HandshakeData
	peerConnWg   *sync.WaitGroup
	peerListener net.Listener
	peerConn     net.Conn
	node         *Node
}

func TestNodeTestSuite(t *testing.T) {
	suite.Run(t, &NodeTestSuite{})
}

func (s *NodeTestSuite) SetupSuite() {
	s.HandshakeData = *CreateHandshakeData(s.T())
}

func setupPeerConnectionForNodeTestSuite(s *NodeTestSuite) {
	var err error
	s.peerListener, err = net.Listen("tcp", s.peerAddr.String())
	if err != nil {
		s.FailNow(err.Error())
	}

	var wg sync.WaitGroup
	s.peerConnWg = &wg
	wg.Add(1)

	go func() {
		defer wg.Done()
		var err error
		s.peerConn, err = s.peerListener.Accept()
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
}

func setupNode(s *NodeTestSuite) {
	s.node = NewNode(
		70015,
		message.NodeNetwork,
		5,
		constants.BlocksFileDirectory,
		20*time.Second,
		10*time.Second,
		10*time.Second,
	)
}

func (s *NodeTestSuite) SetupTest() {
	setupPeerConnectionForNodeTestSuite(s)
	setupNode(s)
}

func (s *NodeTestSuite) TearDownTest() {
	s.peerConnWg.Wait()

	s.peerListener.Close()
	s.peerConn.Close()
}

func (s *NodeTestSuite) TestNode_AddPeerWorks() {
	peer, err := s.node.AddPeer(&s.peerAddr, message.NodeNetwork)
	s.NoError(err)
	s.Equal(1, s.node.peers.Len())
	_, ok := s.node.peers.Get(peer)
	s.True(ok)
}

func (s *NodeTestSuite) TestNode_RemovePeerIfItQuits() {
	peer, err := s.node.AddPeer(&s.peerAddr, message.NodeNetwork)
	s.NoError(err)

	go s.node.Start()

	s.Equal(1, s.node.peers.Len())
	_, ok := s.node.peers.Get(peer)
	s.True(ok)

	// peer has quit
	peer.Quit()
	<-peer.QuitCh

	s.Equal(0, s.node.peers.Len())
	_, ok = s.node.peers.Get(peer)
	s.False(ok)
}

func (s *NodeTestSuite) TestNode_AllPeersQuitIfNodeQuits() {
	peer, err := s.node.AddPeer(&s.peerAddr, message.NodeNetwork)
	s.NoError(err)

	go s.node.Start()

	s.Equal(1, s.node.peers.Len())
	_, ok := s.node.peers.Get(peer)
	s.True(ok)

	// node has quit
	s.node.Quit()

	<-s.node.QuitCh

	s.Equal(0, s.node.peers.Len())
	_, ok = s.node.peers.Get(peer)
	s.False(ok)
}

// TODO - Improve test
func (s *NodeTestSuite) TestNode_PeerRemainsInNodeIfNothingHappens() {
	peer, err := s.node.AddPeer(&s.peerAddr, message.NodeNetwork)
	s.NoError(err)

	go s.node.Start()
	// nothing happens
	time.Sleep(5 * time.Second)

	s.Equal(1, s.node.peers.Len())
	_, ok := s.node.peers.Get(peer)
	s.True(ok)
}
