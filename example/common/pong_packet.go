package common

type PongPacket struct {
}

func (p PongPacket) ID() uint8 {
	return 1
}

func (p *PongPacket) Write(buffer []byte) (int, error) {
	return 0, nil
}

func (p PongPacket) Read(buffer []byte) (int, error) {
	return 0, nil
}
