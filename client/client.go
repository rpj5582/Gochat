package main

import (
	"fmt"
	"time"

	"github.com/rpj5582/Gochat/transport/packet"
)

type ConnectErr struct {
	host string
	err  error
}

func (e ConnectErr) Error() string {
	return fmt.Sprintf("could not connect to %s: %v", e.host, e.err)
}

type NotConnectedErr struct{}

func (e NotConnectedErr) Error() string {
	return "not connected"
}

type DisconnectErr struct{}

func (e DisconnectErr) Error() string {
	return "disconnected from the server"
}

type TimeoutErr struct{}

func (e TimeoutErr) Error() string {
	return "connection to the server timed out"
}

type SendErr struct {
	packetType packet.Type
	err        error
}

func (e SendErr) Error() string {
	return fmt.Sprintf("could not send '%s': %v", e.packetType, e.err)
}

type ReceiveErr struct {
	err error
}

func (e ReceiveErr) Error() string {
	return fmt.Sprintf("could not receive packet: %v", e.err)
}

type Client interface {
	Connect(addr string, timeout time.Duration) error
	Disconnect() error

	SendPacket(packet packet.Packet) error
	ReceivePacket() (packet.Packet, error)
}
