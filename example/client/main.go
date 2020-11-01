package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/rpj5582/gochat/example/common"
	"github.com/rpj5582/gochat/modules/client"
	"github.com/rpj5582/gochat/modules/packet"
)

func main() {
	fmt.Println("welcome to gochat!")
	fmt.Print("Enter a server IP to connect to: ")

	reader := bufio.NewReader(os.Stdin)
	addr, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("error reading server address: %v\n", err)
		return
	}

	addr = strings.Trim(addr, " \t\r\n")

	fmt.Print("Enter a server port to connect to (blank for 20000): ")
	port, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("error reading port: %v\n", err)
		return
	}

	port = strings.Trim(port, " \t\r\n")
	if port == "" {
		port = "20000"
	}

	client := client.NewTCPClient(65535)
	client.RegisterPacketType(&common.PingPacket{}, nil)
	client.RegisterPacketType(&common.PongPacket{}, nil)
	client.RegisterPacketType(&common.MessagePacket{}, func(conn net.Conn, p packet.Packet) {
		messagePacket := p.(*common.MessagePacket)
		fmt.Println(messagePacket.Message)
	})

	if err := client.Connect(addr + ":" + port); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("connected to %s\n", addr+":"+port)

	go func() {
		for {
			err := client.ReceivePacket()
			if err != nil {
				fmt.Println(err)
				client.Disconnect()
				break
			}
		}
	}()

	// Send a test message
	messagePacket := common.MessagePacket{Message: "hello world"}
	if err := client.SendPacket(&messagePacket); err != nil {
		fmt.Println(err)
		client.Disconnect()
		return
	}

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt)
	<-sigChannel
}
