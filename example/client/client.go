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

	client.RegisterPacketType(&shared.MessagePacket{}, func(conn net.Conn, p common.Packet) {
		messagePacket := p.(*shared.MessagePacket)
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
