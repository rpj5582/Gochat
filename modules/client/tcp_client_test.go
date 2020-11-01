package client_test

import (
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/rpj5582/gochat/modules/client"
	"github.com/rpj5582/gochat/modules/common"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/nettest"
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

func TestNewTCPClientInvalidMaxPacketSize(t *testing.T) {
	c, err := client.NewTCPClient(0)
	assert.Nil(t, c)
	assert.IsType(t, &common.InvalidMaxPacketSizeErr{}, err)
}

func TestNewTCPClientSuccess(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func TestTCPClientConnectFailure(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	err = c.Connect("127.0.0.1:20000")
	assert.IsType(t, &client.ConnectErr{}, err)
}

func TestTCPClientConnectSuccess(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	listener, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	err = c.Connect(listener.Addr().String())
	assert.NoError(t, err)
}

func TestTCPClientDisconnectFailure(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	err = c.Disconnect()
	assert.IsType(t, &client.NotConnectedErr{}, err)
}

func TestTCPClientDisconnectSuccess(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	listener, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	err = c.Connect(listener.Addr().String())
	assert.NoError(t, err)

	err = c.Disconnect()
	assert.NoError(t, err)
}

func TestTCPClientAddrNotConnected(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	addr, err := c.Addr()
	assert.Nil(t, addr)
	assert.IsType(t, &client.NotConnectedErr{}, err)
}

func TestTCPClientAddrSuccess(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	listener, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	err = c.Connect(listener.Addr().String())
	assert.NoError(t, err)

	addr, err := c.Addr()
	assert.NotNil(t, addr)
	assert.NoError(t, err)
}

func TestTCPClientServerAddrNotConnected(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	addr, err := c.ServerAddr()
	assert.Nil(t, addr)
	assert.IsType(t, &client.NotConnectedErr{}, err)
}

func TestTCPClientServerAddrSuccess(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	listener, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	err = c.Connect(listener.Addr().String())
	assert.NoError(t, err)

	addr, err := c.ServerAddr()
	assert.NotNil(t, addr)
	assert.NoError(t, err)
}

func TestTCPClientSendPacketNotConnected(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	err = c.SendPacket(&TestPacket{})
	assert.IsType(t, &client.NotConnectedErr{}, err)
}

func TestTCPClientSendPacketFailure(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	listener, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	err = c.Connect(listener.Addr().String())
	assert.NoError(t, err)

	listener.Close()

	err = c.SendPacket(&TestPacket{})
	assert.IsType(t, &common.SendErr{}, err)
}

func TestTCPClientSendPacketSuccess(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	listener, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	err = c.Connect(listener.Addr().String())
	assert.NoError(t, err)

	err = c.SendPacket(&TestPacket{})
	assert.NoError(t, err)
}

func TestTCPClientReceivePacketNotConnected(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	err = c.ReceivePacket()
	assert.IsType(t, &client.NotConnectedErr{}, err)
}

func TestTCPClientReceivePacketDisconnected(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	listener, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	err = c.Connect(listener.Addr().String())
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = c.ReceivePacket()
		assert.IsType(t, &common.DisconnectErr{}, err)
		wg.Done()
	}()

	conn, err := listener.Accept()
	assert.NoError(t, err)

	conn.Close()
	listener.Close()

	wg.Wait()
}

func TestTCPClientReceivePacketNotRegistered(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	listener, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	err = c.Connect(listener.Addr().String())
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = c.ReceivePacket()
		assert.IsType(t, &common.PacketNotRegisteredErr{}, err)
		wg.Done()
	}()

	conn, err := listener.Accept()
	assert.NoError(t, err)

	conn.Write([]byte("unregistered packet"))

	conn.Close()
	listener.Close()

	wg.Wait()
}

func TestTCPClientReceivePacketUnmarshalFailure(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	c.RegisterPacketType(&TestPacket{}, nil)

	listener, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	err = c.Connect(listener.Addr().String())
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = c.ReceivePacket()
		assert.Error(t, err)
		wg.Done()
	}()

	conn, err := listener.Accept()
	assert.NoError(t, err)

	data := []byte{0}
	data = append(data, []byte("unregistered packet")...)
	conn.Write(data)

	conn.Close()
	listener.Close()

	wg.Wait()
}

func TestTCPClientReceivePacketSuccess(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	c.RegisterPacketType(&TestPacket{}, func(conn net.Conn, p common.Packet) {})

	listener, err := nettest.NewLocalListener("tcp")
	assert.NoError(t, err)

	err = c.Connect(listener.Addr().String())
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err = c.ReceivePacket()
		assert.NoError(t, err)
		wg.Done()
	}()

	conn, err := listener.Accept()
	assert.NoError(t, err)

	data := []byte{0}
	data = append(data, []byte("test data")...)
	conn.Write(data)

	conn.Close()
	listener.Close()

	wg.Wait()
}

func TestTCPClientRegisterPacketType(t *testing.T) {
	c, err := client.NewTCPClient(10)
	assert.NotNil(t, c)
	assert.NoError(t, err)

	err = c.RegisterPacketType(&TestPacket{}, func(conn net.Conn, p common.Packet) {})
	assert.NoError(t, err)

	err = c.RegisterPacketType(&TestPacket{}, func(conn net.Conn, p common.Packet) {})
	assert.IsType(t, &common.PacketRegisteredErr{}, err)
}
