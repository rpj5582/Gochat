package client

import (
	"errors"
	"io"
	"net"

	"github.com/rpj5582/gochat/modules/common"
)

// TCPClient is a client that can communicate with a server via TCP
type TCPClient struct {
	registeredPackets map[uint8]struct {
		packet   common.Packet
		callback func(conn net.Conn, p common.Packet)
	}
	maxPacketSize int

	conn        net.Conn
	isConnected bool
}

// NewTCPClient returns an initialized TCP client ready to connect to a server
func NewTCPClient(maxPacketSize int) (*TCPClient, error) {
	if maxPacketSize < 1 {
		return nil, &common.InvalidMaxPacketSizeErr{Size: maxPacketSize}
	}

	return &TCPClient{
		registeredPackets: make(map[uint8]struct {
			packet   common.Packet
			callback func(conn net.Conn, p common.Packet)
		}),
		maxPacketSize: maxPacketSize,
	}, nil
}

func (c *TCPClient) Connect(addr string) error {
	var err error

	if c.conn, err = net.Dial("tcp", addr); err != nil {
		c.conn = nil
		c.isConnected = false
		return &common.ConnectErr{
			Host: addr,
			Err:  err,
		}
	}

	c.isConnected = true
	return nil
}

func (c *TCPClient) Disconnect() error {
	if !c.isConnected {
		return &common.NotConnectedErr{}
	}

	c.conn = nil
	c.isConnected = false
	return nil
}

func (c *TCPClient) Addr() (net.Addr, error) {
	if !c.isConnected {
		return nil, &common.NotConnectedErr{}

	}

	return c.conn.LocalAddr(), nil
}

func (c *TCPClient) ServerAddr() (net.Addr, error) {
	if !c.isConnected {
		return nil, &common.NotConnectedErr{}

	}

	return c.conn.RemoteAddr(), nil
}

func (c *TCPClient) SendPacket(p common.Packet) error {
	if !c.isConnected {
		return &common.NotConnectedErr{}
	}

	packetBuffer := make([]byte, c.maxPacketSize)
	packetID := p.ID()
	packetBuffer[0] = packetID

	n, err := p.Read(packetBuffer[1:])
	if err != nil {
		return err
	}

	if _, err := c.conn.Write(packetBuffer[:n+1]); err != nil {
		return &common.SendErr{
			PacketID: packetID,
			Err:      err,
		}
	}

	return nil
}

func (c *TCPClient) ReceivePacket() error {
	if !c.isConnected {
		return &common.NotConnectedErr{}
	}

	packetBuffer := make([]byte, c.maxPacketSize)
	n, err := c.conn.Read(packetBuffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return &common.TimeoutErr{}
		}

		if err == io.EOF {
			return &common.DisconnectErr{}
		}

		return &common.ReceiveErr{Err: err}
	}

	if n < 1 {
		return &common.ReceiveErr{Err: errors.New("received empty packet")}
	}

	packetID := packetBuffer[0]
	packetBuffer = packetBuffer[1:n]

	p, ok := c.registeredPackets[packetID]
	if !ok {
		return &common.PacketNotRegisteredErr{PacketID: packetID}
	}

	if _, err := p.packet.Write(packetBuffer); err != nil {
		return err
	}

	if p.callback != nil {
		p.callback(c.conn, p.packet)
	}

	return nil
}

func (c *TCPClient) RegisterPacketType(p common.Packet, receiveCallback func(conn net.Conn, p common.Packet)) error {
	packetID := p.ID()
	if _, ok := c.registeredPackets[packetID]; ok {
		return &common.PacketRegisteredErr{PacketID: packetID}
	}

	c.registeredPackets[packetID] = struct {
		packet   common.Packet
		callback func(conn net.Conn, p common.Packet)
	}{packet: p, callback: receiveCallback}

	return nil
}
