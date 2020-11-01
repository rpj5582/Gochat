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

// GLOBALS, DON'T DO THIS
var serv *server.TCPServer
var clientNames map[string]server.ClientID

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

	serv, err = server.NewTCPServer(65535, onClientConnected, onClientDisconnected)
	if err != nil {
		fmt.Println(err)
		return
	}

	clientNames = make(map[string]server.ClientID)

	serv.RegisterPacketType(&shared.ConnectRequest{}, func(clientID server.ClientID, conn net.Conn, p common.Packet) {
		connectionRequest := p.(*shared.ConnectRequest)
		if _, ok := clientNames[connectionRequest.ClientName]; ok {
			fmt.Printf("client attempting to connect with name \"%s\": name already taken\n", connectionRequest.ClientName)
			if err := serv.SendPacket(clientID, &shared.ConnectResponse{Connected: false, ErrMessage: fmt.Sprintf("name \"%s\" is already taken", connectionRequest.ClientName)}); err != nil {
				fmt.Println(err)
				return
			}

			conn.Close()
			return
		}

		clientNames[connectionRequest.ClientName] = clientID
		if err := serv.SendPacket(clientID, &shared.ConnectResponse{Connected: true}); err != nil {
			fmt.Println(err)
			return
		}

		serv.BroadcastPacket(&shared.ConnectedPacket{ClientName: connectionRequest.ClientName}, clientID)
	})

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

func onClientConnected(clientID server.ClientID) {
	fmt.Printf("client connected with ID %d\n", clientID)
}

func onClientDisconnected(clientID server.ClientID, err error) {
	if err != nil {
		fmt.Printf("client with ID %d has disconnected: %v\n", clientID, err)
	}

	fmt.Printf("client with ID %d has disconnected\n", clientID)

	clientName, err := getClientNameFromID(clientNames, clientID)
	if err != nil {
		return
	}

	serv.BroadcastPacket(&shared.DisconnectedPacket{ClientName: clientName}, clientID)
}

func getClientNameFromID(clientNames map[string]server.ClientID, clientID server.ClientID) (string, error) {
	for name, ID := range clientNames {
		if ID == clientID {
			return name, nil
		}
	}

	return "", fmt.Errorf("could not find client name with ID %d", clientID)
}
