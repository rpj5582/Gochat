package server_test

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/rpj5582/gochat/modules/common"
	"github.com/rpj5582/gochat/modules/server"
	"github.com/stretchr/testify/assert"
)

type TestPacket struct{}

func (p TestPacket) ID() uint8 {
	return 0
}

func (p *TestPacket) Write(buffer []byte) (int, error) {
	if buffer[0] != byte('t') {
		return 0, errors.New("write error")
	}
	return len(buffer), nil
}

func (p TestPacket) Read(buffer []byte) (int, error) {
	return copy(buffer, []byte("test data")), nil
}

func TestNewTCPServerInvalidMaxPacketSize(t *testing.T) {
	s, err := server.NewTCPServer(0, nil, nil)
	assert.Nil(t, s)
	assert.IsType(t, &common.InvalidMaxPacketSizeErr{}, err)
}

func TestNewTCPServerSuccess(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)
}

func TestTCPServerStartListenFailure(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	err = s.Start("-1")
	assert.IsType(t, &server.ListenErr{}, err)
}

func TestTCPServerStart(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	var conn net.Conn

	onClientConnected := func(clientID server.ClientID) {
		time.Sleep(time.Millisecond * 10)
		conn.Close()
	}

	onClientDisconnected := func(clientID server.ClientID, err error) {
		assert.NoError(t, err)
		cancelFunc()
	}

	s, err := server.NewTCPServer(10, onClientConnected, onClientDisconnected)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	go func() {
		err = s.Start("0")
		assert.IsType(t, &server.AcceptErr{}, err)
	}()

	time.Sleep(time.Millisecond * 10)

	conn, err = net.Dial("tcp", s.Addr().String())
	assert.NoError(t, err)

	select {
	case <-ctx.Done():
	}

	s.Stop()
}

func TestTCPServerAddrNotListening(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	addr := s.Addr()
	assert.Nil(t, addr)
}

func TestTCPServerAddrSuccess(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	go func() {
		err = s.Start("0")
		assert.IsType(t, &server.AcceptErr{}, err)
	}()

	time.Sleep(time.Millisecond * 10)

	addr := s.Addr()
	assert.NotNil(t, addr)
}
func TestTCPServerSendPacketInvalidClientID(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	err = s.SendPacket(-1, &TestPacket{})
	assert.IsType(t, &server.InvalidClientID{}, err)
}

func TestTCPServerSendPacketFailure(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	clientID := s.AddNewConnection(&net.TCPConn{})

	err = s.SendPacket(clientID, &TestPacket{})
	assert.IsType(t, &common.SendErr{}, err)
}

func TestTCPServerSendPacketSuccess(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	go func() {
		err = s.Start("0")
		assert.IsType(t, &server.AcceptErr{}, err)
	}()

	serverConn, clientConn := net.Pipe()
	clientID := s.AddNewConnection(serverConn)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.SendPacket(clientID, &TestPacket{})
		wg.Done()
	}()

	var buffer [1000]byte
	clientConn.Read(buffer[:])

	wg.Wait()
}

func TestTCPServerBroadcastPacketToAll(t *testing.T) {
	s, err := server.NewTCPServer(10, func(clientID server.ClientID) {}, func(clientID server.ClientID, err error) {})
	assert.NotNil(t, s)
	assert.NoError(t, err)

	err = s.RegisterPacketType(&TestPacket{}, func(clientID server.ClientID, conn net.Conn, p common.Packet) {
		s.BroadcastPacket(p, -1)
	})
	assert.NoError(t, err)

	go func() {
		err = s.Start("0")
		assert.IsType(t, &server.AcceptErr{}, err)
	}()

	time.Sleep(time.Millisecond * 10)

	conn1, err := net.Dial("tcp", s.Addr().String())
	assert.NoError(t, err)

	conn2, err := net.Dial("tcp", s.Addr().String())
	assert.NoError(t, err)

	result1 := []byte{}
	result2 := []byte{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		incoming := []byte{0}
		incoming = append(incoming, []byte("test data")...)
		_, err := conn1.Write(incoming)
		assert.NoError(t, err)

		buffer := [5]byte{}
		n, err := conn1.Read(buffer[:])
		assert.NoError(t, err)

		result1 = buffer[:n]
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		buffer := [5]byte{}
		n, err := conn2.Read(buffer[:])
		assert.NoError(t, err)

		result2 = buffer[:n]
		wg.Done()
	}()

	wg.Wait()

	assert.Equal(t, result1, result2)
}

func TestTCPServerBroadcastPacketExcludeSender(t *testing.T) {
	s, err := server.NewTCPServer(10, func(clientID server.ClientID) {}, func(clientID server.ClientID, err error) {})
	assert.NotNil(t, s)
	assert.NoError(t, err)

	err = s.RegisterPacketType(&TestPacket{}, func(clientID server.ClientID, conn net.Conn, p common.Packet) {
		s.BroadcastPacket(p, clientID)
		conn.Close()
	})
	assert.NoError(t, err)

	go func() {
		err = s.Start("0")
		assert.IsType(t, &server.AcceptErr{}, err)
	}()

	time.Sleep(time.Millisecond * 10)

	conn1, err := net.Dial("tcp", s.Addr().String())
	assert.NoError(t, err)

	conn2, err := net.Dial("tcp", s.Addr().String())
	assert.NoError(t, err)

	result1 := []byte{}
	result2 := []byte{}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		incoming := []byte{0}
		incoming = append(incoming, []byte("test data")...)
		_, err := conn1.Write(incoming)
		assert.NoError(t, err)

		buffer := [5]byte{}
		n, _ := conn1.Read(buffer[:])

		result1 = buffer[:n]
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		buffer := [5]byte{}
		n, err := conn2.Read(buffer[:])
		assert.NoError(t, err)

		result2 = buffer[:n]
		wg.Done()
	}()

	wg.Wait()

	assert.Empty(t, result1)
	assert.NotEmpty(t, result2)
}

func TestTCPServerReceivePacketInvalidClientID(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	err = s.ReceivePacket(-1)
	assert.IsType(t, &server.InvalidClientID{}, err)
}

func TestTCPServerReceivePacketDisconnected(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	serverConn, clientConn := net.Pipe()
	clientID := s.AddNewConnection(serverConn)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = s.ReceivePacket(clientID)
		assert.IsType(t, &common.DisconnectErr{}, err)
		wg.Done()
	}()

	clientConn.Close()

	wg.Wait()
}

func TestTCPServerReceivePacketNotRegistered(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	serverConn, clientConn := net.Pipe()
	clientID := s.AddNewConnection(serverConn)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = s.ReceivePacket(clientID)
		assert.IsType(t, &common.PacketNotRegisteredErr{}, err)
		wg.Done()
	}()

	clientConn.Write([]byte("test data"))

	wg.Wait()
}

func TestTCPServerReceivePacketUnmarshalFailure(t *testing.T) {
	s, err := server.NewTCPServer(15, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	err = s.RegisterPacketType(&TestPacket{}, nil)
	assert.NoError(t, err)

	serverConn, clientConn := net.Pipe()
	clientID := s.AddNewConnection(serverConn)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = s.ReceivePacket(clientID)
		assert.Error(t, err)
		wg.Done()
	}()

	data := []byte{0}
	data = append(data, []byte("unknown data")...)
	clientConn.Write(data)

	wg.Wait()
}

func TestTCPServerReceivePacketSuccess(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	err = s.RegisterPacketType(&TestPacket{}, func(clientID server.ClientID, conn net.Conn, p common.Packet) {})
	assert.NoError(t, err)

	serverConn, clientConn := net.Pipe()
	clientID := s.AddNewConnection(serverConn)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = s.ReceivePacket(clientID)
		assert.NoError(t, err)
		wg.Done()
	}()

	data := []byte{0}
	data = append(data, []byte("test data")...)
	clientConn.Write(data)

	wg.Wait()
}

func TestTCPServerRegisterPacketType(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	err = s.RegisterPacketType(&TestPacket{}, func(clientID server.ClientID, conn net.Conn, p common.Packet) {})
	assert.NoError(t, err)

	err = s.RegisterPacketType(&TestPacket{}, func(clientID server.ClientID, conn net.Conn, p common.Packet) {})
	assert.IsType(t, &common.PacketRegisteredErr{}, err)
}
