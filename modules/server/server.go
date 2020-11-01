package server

import (
	"fmt"
	"net"

	"github.com/rpj5582/gochat/modules/common"
)

// ClientID is a unique ID given to each client upon connection
type ClientID int32

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
	SendPacket(clientID ClientID, p common.Packet) error

	// BroadcastPacket sends a packet to all connected clients, with the option to exclude a client
	BroadcastPacket(p common.Packet, clientIDToExclude ClientID)

	// ReceivePacket receives the next packet from a connection and calls the
	// registered callback function associated with the packet type.
	// This is automatically called when handling a client connection.
	ReceivePacket(clientID ClientID) error

	// RegisterPacketType registers the given packet type as a packet that can be sent and received,
	// and associates a callback function to be called when the given packet type is received.
	// When registering a packet type, pass the zero value for that packet type.
	RegisterPacketType(p common.Packet, receiveCallback func(clientID ClientID, conn net.Conn, p common.Packet)) error
}

// InvalidClientID is an error thrown when a client ID is invalid
type InvalidClientID struct {
	ClientID ClientID
}

func (e *InvalidClientID) Error() string {
	return fmt.Sprintf("invalid client ID of %d", e.ClientID)
}

// ListenErr represents an error encountered when trying to
// listen on a port for incoming connections
type ListenErr struct {
	Port string
	Err  error
}

func (e *ListenErr) Error() string {
	return fmt.Sprintf("failed to listen on port %s: %v", e.Port, e.Err)
}

// AcceptErr represents an error accepting a connection
type AcceptErr struct {
	Err error
}

func (e AcceptErr) Error() string {
	return fmt.Sprintf("could not accept connection: %v", e.Err)
}
