package packet

import (
	"encoding/binary"
	"fmt"
)

type MessagePacket struct {
	Message string
}

func (p MessagePacket) Type() Type {
	return PacketTypeMessage
}

func (p MessagePacket) write(buffer []byte, index *int) error {
	binary.LittleEndian.PutUint16(buffer[*index:], uint16(len(p.Message)))
	*index += 2

	for i := 0; i < len(p.Message); i++ {
		buffer[*index] = p.Message[i]
		*index++
	}

	return nil
}

func (p *MessagePacket) read(buffer []byte, index *int) error {
	messageLength := binary.LittleEndian.Uint16(buffer[*index:])
	*index += 2

	if messageLength > uint16(len(buffer)) {
		return PacketUnmarshalErr{packetType: p.Type(), err: fmt.Errorf("message length %d longer than buffer length %d", messageLength, len(buffer))}
	}

	message := buffer[*index:messageLength]
	p.Message = string(message)
	return nil
}

func (p *MessagePacket) size() int {
	return 4 + len(p.Message)
}
