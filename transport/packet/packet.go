package packet

import (
	"encoding/binary"
	"fmt"
)

type Type int

const (
	PacketTypeUnknown Type = iota
	PacketTypePing
	PacketTypePong
	PacketTypeMessage

	MaxSize = 1024
)

type PacketMarshalErr struct {
	packet Packet
	err    error
}

func (e PacketMarshalErr) Error() string {
	return fmt.Sprintf("failed to marshal packet type %d: %v", e.packet.Type(), e.err)
}

type PacketUnmarshalErr struct {
	packetType Type
	err        error
}

func (e PacketUnmarshalErr) Error() string {
	return fmt.Sprintf("failed to unmarshal packet type %d: %v", e.packetType, e.err)
}

type Packet interface {
	Type() Type

	write(buffer []byte, index *int) error
	read(buffer []byte, index *int) error
	size() int
}

func Marshal(packet Packet) ([]byte, error) {
	data := make([]byte, 2+packet.size())
	var index int

	binary.LittleEndian.PutUint16(data[index:], uint16(packet.Type()))
	index += 2

	err := packet.write(data, &index)
	return data, err
}

func Unmarshal(data []byte) (Packet, error) {
	if len(data) < 2 {
		return nil, PacketUnmarshalErr{packetType: PacketTypeUnknown, err: fmt.Errorf("packet too short (length %d)", len(data))}
	}

	var index int
	packetType := Type(binary.LittleEndian.Uint16(data[index:]))
	index += 2

	switch packetType {
	case PacketTypePing:
		var p PingPacket
		buffer := data[:p.size()]
		return &p, p.read(buffer, &index)

	case PacketTypePong:
		var p PongPacket
		buffer := data[:p.size()]
		return &p, p.read(buffer, &index)

	case PacketTypeMessage:
		var p MessagePacket
		return &p, p.read(data, &index)

	default:
		return nil, PacketUnmarshalErr{packetType: PacketTypeUnknown, err: fmt.Errorf("unknown packet type %d", packetType)}
	}
}
