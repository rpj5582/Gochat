package main

import (
	"errors"
	"io"
	"net"
	"time"

	"github.com/rpj5582/Gochat/transport/packet"
)

type TCPClient struct {
	TimeoutDuration time.Duration
	receiveBuffer   []byte
	conn            net.Conn
	isConnected     bool
}

func (c *TCPClient) Connect(addr string, timeout time.Duration) error {
	var err error
	if c.conn, err = net.DialTimeout("tcp", addr, timeout); err != nil {
		c.conn = nil
		c.isConnected = false
		return ConnectErr{
			host: addr,
			err:  err,
		}
	}

	c.receiveBuffer = make([]byte, packet.MaxSize)
	c.isConnected = true
	c.conn.SetDeadline(time.Now().Add(c.TimeoutDuration))

	return nil
}

func (c *TCPClient) Disconnect() error {
	if !c.isConnected {
		return NotConnectedErr{}
	}

	c.conn = nil
	c.isConnected = false
	return nil
}

func (c *TCPClient) SendPacket(p packet.Packet) error {
	if !c.isConnected {
		return NotConnectedErr{}
	}

	packetBytes, err := packet.Marshal(p)
	if err != nil {
		return err
	}

	if _, err := c.conn.Write(packetBytes); err != nil {
		return SendErr{
			packetType: p.Type(),
			err:        err,
		}
	}

	c.conn.SetDeadline(time.Now().Add(c.TimeoutDuration))
	return nil
}

func (c *TCPClient) ReceivePacket() (packet.Packet, error) {
	if !c.isConnected {
		return nil, NotConnectedErr{}
	}

	n, err := c.conn.Read(c.receiveBuffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, TimeoutErr{}
		}

		if err == io.EOF {
			return nil, DisconnectErr{}
		}

		return nil, ReceiveErr{err: err}
	}

	if n == 0 {
		return nil, ReceiveErr{err: errors.New("received empty packet")}
	}

	if n > packet.MaxSize {
		return nil, ReceiveErr{err: errors.New("packet too large")}
	}

	packet, err := packet.Unmarshal(c.receiveBuffer[:n])
	if err != nil {
		return nil, ReceiveErr{err: err}
	}

	c.conn.SetDeadline(time.Now().Add(c.TimeoutDuration))
	return packet, nil
}
