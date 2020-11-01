package common

import "fmt"

// InvalidMaxPacketSizeErr is an error thrown when the max packet size is invalid
type InvalidMaxPacketSizeErr struct {
	Size int
}

func (e *InvalidMaxPacketSizeErr) Error() string {
	return fmt.Sprintf("invalid max packet size of %d", e.Size)
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

// ConnectErr represents an error establishing a connection
type ConnectErr struct {
	Host string
	Err  error
}

func (e ConnectErr) Error() string {
	return fmt.Sprintf("could not connect to %s: %v", e.Host, e.Err)
}

// NotConnectedErr is returned when a connection is attempting to be used before it is established
type NotConnectedErr struct{}

func (e NotConnectedErr) Error() string {
	return "not connected"
}

// DisconnectErr represents an error that causes a connection to end
type DisconnectErr struct{}

func (e DisconnectErr) Error() string {
	return "disconnected"
}

// TimeoutErr is returned if a connection times out
type TimeoutErr struct{}

func (e TimeoutErr) Error() string {
	return "connection timed out"
}

// SendErr represents an error sending a packet
type SendErr struct {
	PacketID uint8
	Err      error
}

func (e SendErr) Error() string {
	return fmt.Sprintf("could not send packet with ID %d: %v", e.PacketID, e.Err)
}

// ReceiveErr represents an error receiving a packet
type ReceiveErr struct {
	Err error
}

func (e ReceiveErr) Error() string {
	return fmt.Sprintf("could not receive packet: %v", e.Err)
}

// PacketRegisteredErr is returned when a packet is already registered
type PacketRegisteredErr struct {
	PacketID uint8
}

func (e PacketRegisteredErr) Error() string {
	return fmt.Sprintf("packet with ID %d already registered", e.PacketID)
}

// PacketNotRegisteredErr is returned when a packet is being read that has not been registered
type PacketNotRegisteredErr struct {
	PacketID uint8
}

func (e PacketNotRegisteredErr) Error() string {
	return fmt.Sprintf("packet with ID %d has not been registered", e.PacketID)
}
