package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/rpj5582/gochat/example/shared"
	"github.com/rpj5582/gochat/modules/common"
	"github.com/rpj5582/gochat/modules/server"
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

	serv, err := server.NewTCPServer(65535, onClientConnected, onClientDisconnected)
	if err != nil {
		fmt.Println(err)
		return
	}

	serv.RegisterPacketType(&shared.MessagePacket{}, func(clientID server.ClientID, conn net.Conn, p common.Packet) {
		serv.BroadcastPacket(p, clientID)
	})

	go func() {
		if err := serv.Start(port); err != nil {
			fmt.Println(err)
			return
		}
	}()

	fmt.Println("\ngochat server started, waiting for incoming connections")

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt)
	<-sigChannel

	serv.Stop()
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
