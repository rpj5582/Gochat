package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/rpj5582/Gochat/example/common"
	"github.com/rpj5582/Gochat/modules/packet"
	"github.com/rpj5582/Gochat/modules/server"
)

func main() {
	fmt.Print("Enter a port (blank for 20000): ")

	reader := bufio.NewReader(os.Stdin)
	port, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("error reading port: %v\n", err)
		return
	}

	port = strings.Trim(port, " \t\r\n")
	if port == "" {
		port = "20000"
	}

	server := server.NewTCPServer(onClientConnected, onClientDisconnected, 65535)
	server.RegisterPacketType(&common.PingPacket{}, nil)
	server.RegisterPacketType(&common.PongPacket{}, func(conn net.Conn, p packet.Packet) {
		server.SendPacket(conn, p)
	})
	server.RegisterPacketType(&common.MessagePacket{}, func(conn net.Conn, p packet.Packet) {
		server.SendPacket(conn, p)
	})

	go func() {
		if err := server.Start(port); err != nil {
			fmt.Println(err)
			return
		}
	}()

	fmt.Println("\nGochat server started, waiting for incoming connections")

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt)
	<-sigChannel

	server.Stop()
	fmt.Println("\nStopping server")
}

func onClientConnected(clientAddr string) {
	fmt.Printf("client connected from %s\n", clientAddr)
}

func onClientDisconnected(clientAddr string, err error) {
	if err != nil {
		fmt.Printf("client at %s has disconnected: %v\n", clientAddr, err)
	}

	fmt.Printf("client at %s has disconnected\n", clientAddr)
}
