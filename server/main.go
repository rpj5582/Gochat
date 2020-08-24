package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/rpj5582/Gochat/transport/packet"
)

const (
	TimeoutDuration = time.Second * 10
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

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("error listening on port %s: %v\n", port, err)
		return
	}
	defer listener.Close()

	fmt.Println("\nGochat server started, waiting for incoming connections")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("unable to accept connection: %v\n", err)
			continue
		}

		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	conn.SetDeadline(time.Now().Add(TimeoutDuration))
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("client %s connected\n", clientAddr)

	receiveBuffer := make([]byte, packet.MaxSize)
	for {
		if _, err := conn.Read(receiveBuffer); err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Printf("connection to %s closed: client timed out\n", clientAddr)
				return
			}

			if err == io.EOF {
				fmt.Printf("connection to %s closed: client disconnected\n", clientAddr)
				return
			}

			fmt.Printf("closing connection to %s - failed to read incoming data: %v\n", clientAddr, err)
			return
		}

		p, err := packet.Unmarshal(receiveBuffer)
		if err != nil {
			fmt.Printf("closing connection to %s - %v\n", clientAddr, err)
			return
		}

		conn.SetDeadline(time.Now().Add(TimeoutDuration))

		switch p.Type() {
		case packet.PacketTypePing:
			pongPacket := &packet.PongPacket{}
			pongBytes, err := packet.Marshal(pongPacket)
			if err != nil {
				fmt.Printf("closing connection to %s - failed to marshal pong packet: %v\n", clientAddr, err)
				return
			}

			if _, err := conn.Write(pongBytes); err != nil {
				fmt.Printf("closing connection to %s - failed to send pong packet: %v\n", clientAddr, err)
				return
			}

		case packet.PacketTypeMessage:
			messageBytes, err := packet.Marshal(p)
			if err != nil {
				fmt.Printf("closing connection to %s - failed to marshal message packet: %v\n", clientAddr, err)
				return
			}

			if _, err := conn.Write(messageBytes); err != nil {
				fmt.Printf("closing connection to %s - failed to send message packet: %v\n", clientAddr, err)
				return
			}
		}
	}
}
