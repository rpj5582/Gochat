package client

import (
	"net"

	"github.com/rpj5582/gochat/modules/common"
)

// Client represents a network client that can communicate with a server,
// and by proxy, other clients the server is connected to
type Client interface {
	// Connect connects a client to a server at the given address
	Connect(addr string) error

	// Disconnect disconnects the client from the server
	Disconnect() error

	// SendPacket sends the given packet to the server
	SendPacket(p common.Packet) error

	// ReceivePacket receives the next packet from the server and calls the
	// registered callback function associated with the packet type
	ReceivePacket() error

	// RegisterPacketType registers the given packet type as a packet that can be sent and received,
	// and associates a callback function to be called when the given packet type is received.
	// When registering a packet type, pass the zero value for that packet type.
	RegisterPacketType(p common.Packet, receiveCallback func(conn net.Conn, p common.Packet)) error
}
