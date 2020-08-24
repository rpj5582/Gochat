package packet

type PongPacket struct {
}

func (p PongPacket) Type() Type {
	return PacketTypePong
}

func (p PongPacket) write(buffer []byte, index *int) error {
	return nil
}

func (p *PongPacket) read(buffer []byte, index *int) error {
	return nil
}

func (p *PongPacket) size() int {
	return 0
}
