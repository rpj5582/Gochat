package shared

type PingPacket struct {
}

func (p PingPacket) ID() uint8 {
	return 0
}

func (p *PingPacket) Write(buffer []byte) (int, error) {
	return 0, nil
}

func (p PingPacket) Read(buffer []byte) (int, error) {
	return 0, nil
}
