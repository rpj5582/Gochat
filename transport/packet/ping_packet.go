package packet

type PingPacket struct {
}

func (p PingPacket) Type() Type {
	return PacketTypePing
}

func (p PingPacket) write(buffer []byte, index *int) error {
	return nil
}

func (p *PingPacket) read(buffer []byte, index *int) error {
	return nil
}

func (p *PingPacket) size() int {
	return 0
}
