package shared

import (
	"encoding/binary"
	"fmt"
)

// ConnectRequest implements the Packet interface and is used to ask the server to connect
type ConnectRequest struct {
	ClientName string
}

func (p ConnectRequest) ID() uint8 {
	return ConnectRequestPacketID
}

func (p *ConnectRequest) Write(buffer []byte) (int, error) {
	nameLength := binary.LittleEndian.Uint16(buffer)

	if nameLength > uint16(len(buffer)) {
		return 2, fmt.Errorf("failed to write connect request packet: name length %d longer than buffer length %d", nameLength, len(buffer))
	}

	name := buffer[2 : 2+nameLength]
	p.ClientName = string(name)
	return 2 + len(p.ClientName), nil
}

func (p ConnectRequest) Read(buffer []byte) (int, error) {
	var index int

	binary.LittleEndian.PutUint16(buffer, uint16(len(p.ClientName)))
	index += 2

	for i := 0; i < len(p.ClientName); i++ {
		buffer[index] = p.ClientName[i]
		index++
	}

	return index, nil
}

// ConnectResponse implements the Packet interface and is used by the server to
// tell a client if they were allowed to connect
type ConnectResponse struct {
	Connected  bool
	ErrMessage string
}

func (p ConnectResponse) ID() uint8 {
	return ConnectResponsePacketID
}

func (p *ConnectResponse) Write(buffer []byte) (int, error) {
	var index int

	if len(buffer) == 0 {
		return 0, nil
	}

	connected := buffer[0] == 1
	index++

	if connected {
		p.Connected = true
		return index, nil
	}

	errMessageLength := binary.LittleEndian.Uint16(buffer[index:])
	index += 2

	if errMessageLength > uint16(len(buffer)) {
		return index, fmt.Errorf("failed to write connect response packet: error message length %d longer than buffer length %d", errMessageLength, len(buffer))
	}

	errMessage := buffer[index : index+int(errMessageLength)]
	p.ErrMessage = string(errMessage)
	return index + len(p.ErrMessage), nil
}

func (p ConnectResponse) Read(buffer []byte) (int, error) {
	var index int

	if len(buffer) == 0 {
		return 0, nil
	}

	if p.Connected {
		buffer[index] = 1
		index++
		return index, nil
	}

	buffer[index] = 0
	index++

	binary.LittleEndian.PutUint16(buffer[index:], uint16(len(p.ErrMessage)))
	index += 2

	for i := 0; i < len(p.ErrMessage); i++ {
		buffer[index] = p.ErrMessage[i]
		index++
	}

	return index, nil
}

// ConnectedPacket implements the Packet interface and is used to inform other clients that a client has connected
type ConnectedPacket struct {
	ClientName string
}

func (p ConnectedPacket) ID() uint8 {
	return ConnectedPacketID
}

func (p *ConnectedPacket) Write(buffer []byte) (int, error) {
	nameLength := binary.LittleEndian.Uint16(buffer)

	if nameLength > uint16(len(buffer)) {
		return 2, fmt.Errorf("failed to write connected packet: name length %d longer than buffer length %d", nameLength, len(buffer))
	}

	name := buffer[2 : 2+nameLength]
	p.ClientName = string(name)
	return 2 + len(p.ClientName), nil
}

func (p ConnectedPacket) Read(buffer []byte) (int, error) {
	var index int

	binary.LittleEndian.PutUint16(buffer, uint16(len(p.ClientName)))
	index += 2

	for i := 0; i < len(p.ClientName); i++ {
		buffer[index] = p.ClientName[i]
		index++
	}

	return index, nil
}

// DisconnectedPacket implements the Packet interface and is used to inform other clients that a client has disconnected
type DisconnectedPacket struct {
	ClientName string
}

func (p DisconnectedPacket) ID() uint8 {
	return DisconnectedPacketID
}

func (p *DisconnectedPacket) Write(buffer []byte) (int, error) {
	nameLength := binary.LittleEndian.Uint16(buffer)

	if nameLength > uint16(len(buffer)) {
		return 2, fmt.Errorf("failed to write disconnected packet: name length %d longer than buffer length %d", nameLength, len(buffer))
	}

	name := buffer[2 : 2+nameLength]
	p.ClientName = string(name)
	return 2 + len(p.ClientName), nil
}

func (p DisconnectedPacket) Read(buffer []byte) (int, error) {
	var index int

	binary.LittleEndian.PutUint16(buffer, uint16(len(p.ClientName)))
	index += 2

	for i := 0; i < len(p.ClientName); i++ {
		buffer[index] = p.ClientName[i]
		index++
	}

	return index, nil
}
