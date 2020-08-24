package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rpj5582/Gochat/transport/packet"
)

func main() {
	fmt.Println("welcome to Gochat!")
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

	client := TCPClient{TimeoutDuration: time.Second * 10}
	if err := client.Connect(addr+":"+port, time.Second*10); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("connected to %s\n", addr+":"+port)

	// heartbeat loop
	go func() {
		pingPacket := packet.PingPacket{}
		for {
			if err := client.SendPacket(&pingPacket); err != nil {
				fmt.Println(err)
				client.Disconnect()
				return
			}

			_, err := client.ReceivePacket()
			if err != nil {
				fmt.Println(err)
				client.Disconnect()
				return
			}

			time.Sleep(time.Second)
		}
	}()

	// Send a test message
	messagePacket := packet.MessagePacket{Message: "hello world"}
	if err := client.SendPacket(&messagePacket); err != nil {
		fmt.Println(err)
		client.Disconnect()
		return
	}

	p, err := client.ReceivePacket()
	if err != nil {
		fmt.Println(err)
		client.Disconnect()
		return
	}

	switch p.Type() {
	case packet.PacketTypeMessage:
		fmt.Println(messagePacket.Message)
	}
}
