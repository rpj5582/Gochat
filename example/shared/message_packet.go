package shared

import (
	"encoding/binary"
	"fmt"
)

// MessagePacket implements the Packet interface and carries a single message across the network
type MessagePacket struct {
	Message string
}

func (p MessagePacket) ID() uint8 {
	return MessagePacketID
}

func (p *MessagePacket) Write(buffer []byte) (int, error) {
	messageLength := binary.LittleEndian.Uint16(buffer)

	if messageLength > uint16(len(buffer)) {
		return 2, fmt.Errorf("failed to read message packet: message length %d longer than buffer length %d", messageLength, len(buffer))
	}

	message := buffer[2 : 2+messageLength]
	p.Message = string(message)
	return 2 + len(p.Message), nil
}

func (p MessagePacket) Read(buffer []byte) (int, error) {
	var index int

	binary.LittleEndian.PutUint16(buffer, uint16(len(p.Message)))
	index += 2

	for i := 0; i < len(p.Message); i++ {
		buffer[index] = p.Message[i]
		index++
	}

	return index, nil
}
