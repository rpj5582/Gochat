package server

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/rpj5582/gochat/modules/common"
)

// TCPServer is a server that can communicate with clients via TCP
type TCPServer struct {
	registeredPackets map[uint8]struct {
		packet   common.Packet
		callback func(clientID ClientID, conn net.Conn, p common.Packet)
	}
	maxPacketSize int

	listener    net.Listener
	connections map[ClientID]net.Conn
	connMutex   sync.RWMutex

	onClientConnected    func(clientID ClientID)
	onClientDisconnected func(clientID ClientID, err error)

	clientCounter ClientID
}

// NewTCPServer returns an initialized TCP server ready to start
// listening for incoming client connections
func NewTCPServer(maxPacketSize int, onClientConnected func(clientID ClientID), onClientDisconnected func(clientID ClientID, err error)) (*TCPServer, error) {
	if maxPacketSize < 1 {
		return nil, &common.InvalidMaxPacketSizeErr{Size: maxPacketSize}
	}

	return &TCPServer{
		registeredPackets: make(map[uint8]struct {
			packet   common.Packet
			callback func(clientID ClientID, conn net.Conn, p common.Packet)
		}),
		maxPacketSize:        maxPacketSize,
		connections:          make(map[ClientID]net.Conn),
		onClientConnected:    onClientConnected,
		onClientDisconnected: onClientDisconnected,
	}, nil
}

func (s *TCPServer) Start(port string) error {
	var err error
	s.listener, err = net.Listen("tcp", ":"+port)
	if err != nil {
		return &ListenErr{Port: port, Err: err}
	}

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return &AcceptErr{Err: err}
		}

		clientID := s.AddNewConnection(conn)

		go func(clientID ClientID, conn net.Conn) {
			defer func() {
				conn.Close()
				s.connMutex.Lock()
				delete(s.connections, clientID)
				s.connMutex.Unlock()
			}()

			s.onClientConnected(clientID)

			var err error
			for {
				if err = s.ReceivePacket(clientID); err != nil {
					switch err.(type) {
					case *common.DisconnectErr:
						err = nil
					}

					break
				}
			}

			s.onClientDisconnected(clientID, err)
		}(clientID, conn)
	}
}

func (s *TCPServer) AddNewConnection(conn net.Conn) ClientID {
	s.connMutex.Lock()
	clientID := s.clientCounter
	s.clientCounter++
	s.connections[clientID] = conn
	s.connMutex.Unlock()

	return clientID
}

func (s *TCPServer) Stop() {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	for _, conn := range s.connections {
		conn.Close()
	}

	s.connections = nil
	s.listener.Close()
}

func (s *TCPServer) Addr() net.Addr {
	if s.listener != nil {
		return s.listener.Addr()
	}

	return nil
}

func (s *TCPServer) SendPacket(clientID ClientID, p common.Packet) error {
	packetBuffer := make([]byte, s.maxPacketSize)
	packetBuffer[0] = p.ID()

	n, err := p.Read(packetBuffer[1:])
	if err != nil {
		return err
	}

	s.connMutex.RLock()
	conn, ok := s.connections[clientID]
	if !ok {
		s.connMutex.RUnlock()
		return &InvalidClientID{ClientID: clientID}
	}
	s.connMutex.RUnlock()

	if _, err := conn.Write(packetBuffer[:n+1]); err != nil {
		return &common.SendErr{
			PacketID: p.ID(),
			Err:      err,
		}
	}

	return nil
}

func (s *TCPServer) BroadcastPacket(p common.Packet, clientIDToExclude ClientID) {
	s.connMutex.RLock()
	defer s.connMutex.RUnlock()

	for clientID := range s.connections {
		if clientID != clientIDToExclude {
			s.SendPacket(clientID, p)
		}
	}
}

func (s *TCPServer) ReceivePacket(clientID ClientID) error {
	packetBuffer := make([]byte, s.maxPacketSize)

	s.connMutex.RLock()
	conn, ok := s.connections[clientID]
	if !ok {
		s.connMutex.RUnlock()
		return &InvalidClientID{ClientID: clientID}
	}
	s.connMutex.RUnlock()

	n, err := conn.Read(packetBuffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return &common.TimeoutErr{}
		}

		if err == io.EOF {
			return &common.DisconnectErr{}
		}

		return &common.ReceiveErr{Err: err}
	}

	if n < 1 {
		return &common.ReceiveErr{Err: errors.New("received empty packet")}
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

	p.callback(clientID, conn, p.packet)
	return nil
}

func (s *TCPServer) RegisterPacketType(p common.Packet, receiveCallback func(clientID ClientID, conn net.Conn, p common.Packet)) error {
	packetID := p.ID()
	if _, ok := s.registeredPackets[packetID]; ok {
		return &common.PacketRegisteredErr{PacketID: packetID}
	}

	s.registeredPackets[packetID] = struct {
		packet   common.Packet
		callback func(clientID ClientID, conn net.Conn, p common.Packet)
	}{packet: p, callback: receiveCallback}

	return nil
}
