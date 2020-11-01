package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/rpj5582/gochat/example/shared"
	"github.com/rpj5582/gochat/modules/client"
	"github.com/rpj5582/gochat/modules/common"
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

	client, err := client.NewTCPClient(65535)
	if err != nil {
		fmt.Println(err)
		return
	}

	client.RegisterPacketType(&shared.ConnectResponse{}, func(conn net.Conn, p common.Packet) {
		connectedResponse := p.(*shared.ConnectResponse)
		if !connectedResponse.Connected {
			fmt.Printf("server rejected connection request: %s\n", connectedResponse.ErrMessage)
			conn.Close()
			return
		}

		fmt.Printf("You have joined the chat\n")
	})

	client.RegisterPacketType(&shared.ConnectedPacket{}, func(conn net.Conn, p common.Packet) {
		connectedPacket := p.(*shared.ConnectedPacket)
		fmt.Printf("%s has join the chat\n", connectedPacket.ClientName)
	})

	client.RegisterPacketType(&shared.MessagePacket{}, func(conn net.Conn, p common.Packet) {
		messagePacket := p.(*shared.MessagePacket)
		fmt.Println(messagePacket.Message)
	})

	client.RegisterPacketType(&shared.DisconnectedPacket{}, func(conn net.Conn, p common.Packet) {
		disconnectPacket := p.(*shared.DisconnectedPacket)
		fmt.Printf("%s has left the chat\n", disconnectPacket.ClientName)
	})

	if err := client.Connect(addr + ":" + port); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("connected to %s\n", addr+":"+port)

	if err := client.SendPacket(&shared.ConnectRequest{ClientName: "Devin"}); err != nil {
		fmt.Println(err)
		return
	}

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

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading input message: %v\n", err)
			return
		}

		message = strings.Trim(message, " \t\r\n")
		if message == "" {
			continue
		}

		messagePacket := shared.MessagePacket{Message: message}
		if err := client.SendPacket(&messagePacket); err != nil {
			fmt.Println(err)
			client.Disconnect()
			return
		}
	}
}
