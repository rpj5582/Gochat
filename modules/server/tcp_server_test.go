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
	assert.IsType(t, &common.ListenErr{}, err)
}

func TestTCPServerStart(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	var conn net.Conn

	onClientConnected := func(clientAddr string) {
		time.Sleep(time.Millisecond * 10)
		conn.Close()
	}

	onClientDisconnected := func(clientAddr string, err error) {
		assert.NoError(t, err)
		cancelFunc()
	}

	s, err := server.NewTCPServer(10, onClientConnected, onClientDisconnected)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	go func() {
		err = s.Start("0")
		assert.IsType(t, &common.AcceptErr{}, err)
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
		assert.IsType(t, &common.AcceptErr{}, err)
	}()

	time.Sleep(time.Millisecond * 10)

	addr := s.Addr()
	assert.NotNil(t, addr)
}
func TestTCPServerSendPacketFailure(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	s.SendPacket(&net.TCPConn{}, &TestPacket{})
}

func TestTCPServerSendPacketSuccess(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	go func() {
		err = s.Start("0")
		assert.IsType(t, &common.AcceptErr{}, err)
	}()

	serverConn, clientConn := net.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.SendPacket(serverConn, &TestPacket{})
		wg.Done()
	}()

	var buffer [1000]byte
	clientConn.Read(buffer[:])

	wg.Wait()
}

func TestTCPServerReceivePacketDisconnected(t *testing.T) {
	s, err := server.NewTCPServer(10, nil, nil)
	assert.NotNil(t, s)
	assert.NoError(t, err)

	serverConn, clientConn := net.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = s.ReceivePacket(serverConn)
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

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = s.ReceivePacket(serverConn)
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

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = s.ReceivePacket(serverConn)
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

	err = s.RegisterPacketType(&TestPacket{}, func(conn net.Conn, p common.Packet) {})
	assert.NoError(t, err)

	serverConn, clientConn := net.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = s.ReceivePacket(serverConn)
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

	err = s.RegisterPacketType(&TestPacket{}, func(conn net.Conn, p common.Packet) {})
	assert.NoError(t, err)

	err = s.RegisterPacketType(&TestPacket{}, func(conn net.Conn, p common.Packet) {})
	assert.IsType(t, &common.PacketRegisteredErr{}, err)
}
