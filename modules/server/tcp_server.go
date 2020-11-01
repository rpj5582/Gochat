package server

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/rpj5582/Gochat/modules/common"
	"github.com/rpj5582/Gochat/modules/packet"
)

// TCPServer is a server that can communicate with clients via TCP
type TCPServer struct {
	registeredPackets map[uint8]struct {
		packet   packet.Packet
		callback func(conn net.Conn, p packet.Packet)
	}
	maxPacketSize int

	listener    net.Listener
	connections map[net.Conn]struct{}
	connMutex   sync.Mutex

	onClientConnected    func(clientAddr string)
	onClientDisconnected func(clientAddr string, err error)
}

// NewTCPServer returns an initialized TCP server ready to start
// listening for incoming client connections
func NewTCPServer(onClientConnected func(clientAddr string), onClientDisconnected func(clientAddr string, err error), maxPacketSize int) *TCPServer {
	return &TCPServer{
		registeredPackets: make(map[uint8]struct {
			packet   packet.Packet
			callback func(conn net.Conn, p packet.Packet)
		}),
		maxPacketSize:        maxPacketSize,
		connections:          make(map[net.Conn]struct{}),
		onClientConnected:    onClientConnected,
		onClientDisconnected: onClientDisconnected,
	}
}

func (s *TCPServer) Start(port string) error {
	var err error
	s.listener, err = net.Listen("tcp", ":"+port)
	if err != nil {
		return &common.ListenErr{Port: port, Err: err}
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return &common.AcceptErr{Err: err}
		}

		s.connMutex.Lock()
		s.connections[conn] = struct{}{}
		s.connMutex.Unlock()

		go func() {
			defer func() {
				conn.Close()
				s.connMutex.Lock()
				delete(s.connections, conn)
				s.connMutex.Unlock()
			}()

			remoteAddr := conn.RemoteAddr().String()

			s.onClientConnected(remoteAddr)

			var err error
			for {
				if err = s.ReceivePacket(conn); err != nil {
					switch err.(type) {
					case common.DisconnectErr:
						err = nil
					}

					break
				}
			}

			s.onClientDisconnected(remoteAddr, err)
		}()
	}
}

func (s *TCPServer) Stop() {
	for conn := range s.connections {
		conn.Close()
	}

	s.connections = nil
	s.listener.Close()
}

func (s *TCPServer) SendPacket(conn net.Conn, p packet.Packet) error {
	packetBuffer := make([]byte, s.maxPacketSize)
	packetBuffer[0] = p.ID()

	n, err := p.Read(packetBuffer[1:])
	if err != nil {
		return err
	}

	if _, err := conn.Write(packetBuffer[:n+1]); err != nil {
		return common.SendErr{
			PacketID: p.ID(),
			Err:      err,
		}
	}

	return nil
}

func (s *TCPServer) ReceivePacket(conn net.Conn) error {
	packetBuffer := make([]byte, s.maxPacketSize)
	n, err := conn.Read(packetBuffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return &common.TimeoutErr{}
		}

		if err == io.EOF {
			return common.DisconnectErr{}
		}

		return common.ReceiveErr{Err: err}
	}

	if n < 1 {
		return common.ReceiveErr{Err: errors.New("received empty packet")}
	}

	packetID := packetBuffer[0]
	packetBuffer = packetBuffer[1:n]

	p, ok := s.registeredPackets[packetID]
	if !ok {
		return &common.PacketNotRegisteredErr{PacketID: packetID}
	}

	if _, err := p.packet.Write(packetBuffer); err != nil {
		return err
	}

	p.callback(conn, p.packet)
	return nil
}

func (s *TCPServer) RegisterPacketType(p packet.Packet, receiveCallback func(conn net.Conn, p packet.Packet)) error {
	packetID := p.ID()
	if _, ok := s.registeredPackets[packetID]; ok {
		return &common.PacketRegisteredErr{PacketID: packetID}
	}

	s.registeredPackets[packetID] = struct {
		packet   packet.Packet
		callback func(conn net.Conn, p packet.Packet)
	}{packet: p, callback: receiveCallback}

	return nil
}
