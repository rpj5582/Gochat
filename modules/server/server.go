package server

import (
	"net"

	"github.com/rpj5582/gochat/modules/common"
)

// Server is an interface for a network server that can accept
// connections from and communicate with clients
type Server interface {
	// Start establishes the server and begins listening for client connections
	Start(port int) error

	// Stop disconnects all clients and shuts down the server
	Stop()

	// Addr returns the address of the server
	Addr() net.Addr

	// SendPacket sends the given packet to a given connection
	SendPacket(conn net.Conn, p common.Packet) error

	// BroadcastPacket sends a packet to all connected clients
	BroadcastPacket(p common.Packet)

	// ReceivePacket receives the next packet from a connection and calls the
	// registered callback function associated with the packet type.
	// This is automatically called when handling a client connection.
	ReceivePacket(conn net.Conn) error

	// RegisterPacketType registers the given packet type as a packet that can be sent and received,
	// and associates a callback function to be called when the given packet type is received.
	// When registering a packet type, pass the zero value for that packet type.
	RegisterPacketType(p common.Packet, receiveCallback func(conn net.Conn, p common.Packet)) error
}
