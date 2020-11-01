package shared

const (
	// MaxPacketSize is the maximum size of a packet in bytes
	MaxPacketSize = 1<<16 - 1

	// MaxNameLength is the maximum length of a client's name
	MaxNameLength = 32
)

const (
	UnknownPacketID = iota
	MessagePacketID
	ConnectRequestPacketID
	ConnectResponsePacketID
	ConnectedPacketID
	DisconnectedPacketID
)
