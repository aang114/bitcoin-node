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

func exchangeVersionMessage(conn *net.TCPConn, services message.Services, receivingServices message.Services) error {
	localTcpAddr, err := getLocalAddr(conn)
	if err != nil {
		return err
	}
	remoteTcpAddr, err := getRemoteAddr(conn)
	if err != nil {
		return err
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

	// receive version message
	msg, err = message.DecodeMessage(conn)
	if err != nil {
		return err
	}
	if msg.Header.Command != message.VersionCommand {
		// TODO - improve error
		return errors.New("invalid Command")
	}
	if msg.Header.Magic != constants.MainnetMagicValue {
		// TODO error
		return errors.New("invalid Magic")
	}

	payload, ok := msg.Payload.(*message.VersionPayload)
	if !ok {
		return errors.New("invalid Payload")
	}

	if payload.Version > constants.ProtocolVersion {
		// TODO error
		return errors.New("protocol version not supported")
	}

	log.Printf("exchanged version message between localAddr (%+v) and remoteAddr (%+v)\n", localTcpAddr, remoteTcpAddr)

	return nil
}

func exchangeVerackMessage(conn *net.TCPConn) error {
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
	if msg.Header.Command != message.VerackCommand {
		return errors.New("invalid Command")
	}
	if msg.Header.Magic != constants.MainnetMagicValue {
		// TODO error
		return errors.New("invalid Magic")
	}

	log.Println("exchanged verack message")

	return nil
}

func PerformHandshake(remoteAddr net.TCPAddr, services message.Services, receivingServices message.Services) (*net.TCPConn, error) {
	log.Printf("performing handshake with remoteAddr %+v\n", remoteAddr)
	conn, err := net.DialTCP("tcp", nil, &remoteAddr)
	if err != nil {
		return nil, err
	}

	err = exchangeVersionMessage(conn, services, receivingServices)
	if err != nil {
		return nil, err
	}
	err = exchangeVerackMessage(conn)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
