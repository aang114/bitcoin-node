package networking

import (
	"errors"
	"github.com/aang114/bitcoin-node/constants"
	"github.com/aang114/bitcoin-node/message"
	"log"
	"math/rand"
	"net"
	"time"
)

func getLocalAddr(conn *net.TCPConn) (*net.TCPAddr, error) {
	localTcpAddr, ok := conn.LocalAddr().(*net.TCPAddr)
	if !ok {
		return nil, errors.New("local address is not a tcp address")
	}
	return localTcpAddr, nil
}

func getRemoteAddr(conn *net.TCPConn) (*net.TCPAddr, error) {
	remoteTcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		return nil, errors.New("remote address is not a tcp address")
	}
	return remoteTcpAddr, nil
}

func exchangeVersionMessage(conn *net.TCPConn, services message.Services, receivingServices message.Services) (*message.VersionPayload, error) {
	localTcpAddr, err := getLocalAddr(conn)
	if err != nil {
		return nil, err
	}
	remoteTcpAddr, err := getRemoteAddr(conn)
	if err != nil {
		return nil, err
	}

	// send version message
	msg, err := message.NewVersionMessage(
		constants.ProtocolVersion,
		message.NodeNetwork,
		time.Now().Unix(),
		*message.NewNetworkAddress(receivingServices, remoteTcpAddr.IP, uint16(remoteTcpAddr.Port)),
		*message.NewNetworkAddress(services, localTcpAddr.IP, uint16(localTcpAddr.Port)),
		rand.Uint64(),
		constants.UserAgent,
		0,
		false)
	if err != nil {
		return nil, err
	}
	encoded, err := msg.Encode()
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(encoded)
	if err != nil {
		return nil, err
	}

	// receive version message
	msg, err = message.DecodeMessage(conn)
	if err != nil {
		return nil, err
	}
	if msg.Header.Command != message.VersionCommand {
		return nil, errors.New("invalid Command")
	}
	if msg.Header.Magic != constants.MainnetMagicValue {
		return nil, errors.New("invalid Magic")
	}

	payload, ok := msg.Payload.(*message.VersionPayload)
	if !ok {
		return nil, errors.New("invalid Payload")
	}

	if payload.Version > constants.ProtocolVersion {
		return nil, errors.New("protocol version not supported")
	}

	log.Printf("🔄 Exchanged version message with peer %s", conn.RemoteAddr())

	return payload, nil
}

func exchangeVerackMessage(conn *net.TCPConn, receivedVersionNumber int32) error {
	// send verack message
	msg, err := message.NewVerackMessage()
	if err != nil {
		return err
	}
	encoded, err := msg.Encode()
	if err != nil {
		return err
	}
	_, err = conn.Write(encoded)
	if err != nil {
		return err
	}

	// receive verack message
	msg, err = message.DecodeMessage(conn)
	if err != nil {
		return err
	}
	if receivedVersionNumber >= 70016 {
		if msg.Header.Magic != constants.MainnetMagicValue {
			return errors.New("invalid Magic")
		}
		// Before receiving a VERACK, a node should not send anything but VERSION/VERACK and feature negotiation messages (WTXIDRELAY, SENDADDRV2). (https://github.com/bitcoin/bitcoin/blob/e9262ea32a6e1d364fb7974844fadc36f931f8c6/test/functional/p2p_leak.py#L7-L8)
		if msg.Header.Command == message.SendAddrV2Command {
			msg, err = message.DecodeMessage(conn)
			if err != nil {
				return err
			}
		}
	}
	if msg.Header.Command != message.VerackCommand {
		return errors.New("invalid Command")
	}
	if msg.Header.Magic != constants.MainnetMagicValue {
		return errors.New("invalid Magic")
	}

	log.Printf("🔄 Exchanged verack message with peer %s", conn.RemoteAddr())

	return nil
}

func exchangeWtxidrelayMessage(conn *net.TCPConn) error {
	// send wtxidrelay message
	msg, err := message.NewWtxidRelayMessage()
	if err != nil {
		return err
	}
	encoded, err := msg.Encode()
	if err != nil {
		return err
	}
	_, err = conn.Write(encoded)
	if err != nil {
		return err
	}

	// receive wtxidrelay message
	msg, err = message.DecodeMessage(conn)
	if err != nil {
		return err
	}
	if msg.Header.Command != message.WtxidRelayCommand {
		return errors.New("invalid Command")
	}
	if msg.Header.Magic != constants.MainnetMagicValue {
		return errors.New("invalid Magic")
	}

	log.Printf("🔄 Exchanged wtxidrelay message with peer %s", conn.RemoteAddr())

	return nil
}

func PerformHandshake(remoteAddr *net.TCPAddr, tcpTimeout time.Duration, services message.Services, receivingServices message.Services) (*net.TCPConn, error) {
	log.Printf("🤝 Performing handshake with peer %s", remoteAddr.String())
	//conn, err := net.DialTCP("tcp", nil, &remoteAddr)
	connI, err := net.DialTimeout("tcp", remoteAddr.String(), tcpTimeout)
	if err != nil {
		return nil, err
	}
	conn, ok := connI.(*net.TCPConn)
	if !ok {
		return nil, errors.New("Could not convert net.Conn to *net.TCPConn")
	}
	receivedVersionPayload, err := exchangeVersionMessage(conn, services, receivingServices)
	if err != nil {
		return nil, err
	}
	// The wtxidrelay message MUST be sent in response to a version message from a peer whose protocol version is >= 70016 and prior to sending a verack. A wtxidrelay message received after a verack message MUST be ignored or treated as invalid. (https://bips.dev/339/)
	if receivedVersionPayload.Version >= 70016 {
		err = exchangeWtxidrelayMessage(conn)
		if err != nil {
			return nil, err
		}
	}
	err = exchangeVerackMessage(conn, receivedVersionPayload.Version)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Handshake successful with peer %s!", conn.RemoteAddr())

	return conn, nil
}
