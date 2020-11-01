package common

import (
	"io"
)

// Packet is the interface a client and server use to send packets across the network
type Packet interface {
	io.ReadWriter

	ID() uint8
}
