// socket-server project main.go
package sockets

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	clientsFname = "clients.txt"
	serverType   = "tcp"
)

type Data struct {
	clientIDs []string // A list of client IDs
}

// idExists checks if the given client ID string exists in the
// Data.clientIDs slice of strings and returns true if it is found.
func (d *Data) idExists(newClientID string) bool {
	for _, existingClientID := range d.clientIDs {
		if newClientID == existingClientID {
			return true
		}
	}
	return false
}

// NewData creates a new Data struct and returns it's pointer.
func NewData() *Data {
	return &Data{
		clientIDs: []string{},
	}
}

// Server establishes a TCP server using network sockets capable of receiving
// messages from multiple clients. Using net.Listen Server listens for new
// clients, creates a new client session on a new goroutine, and passes them a
// pointer to a shared Data structure so they can be processed.
func Server(serverHost, serverPort string) {
	// Open server and close upon function completion.
	fmt.Println("Server Running...")
	server, err := net.Listen(serverType, serverHost+":"+serverPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()

	// Begin listening for clients and establishing sessions.
	fmt.Println("Listening on " + serverHost + ":" + serverPort)
	fmt.Println("Waiting for client...")
	data := NewData()
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("client connected")
		go clientSession(connection, data)
	}
}

// clientSession handles a client connection by repeatedly reading the
// connection buffer and searching for keyword prefixes.
// "CONNECT client_id" will close the connection if the given ID exists. If not,
// add the ID to the Data structure.
func clientSession(connection net.Conn, data *Data) {
	defer connection.Close()
	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}
		switch {
		// No message
		case mLen == 0:
			continue
		// CONNECT
		case strings.HasPrefix(string(buffer[:mLen]), "CONNECT "):
			id := string(buffer[8:mLen])
			if data.idExists(id) {
				_, err = connection.Write([]byte("CONNECT: ERROR"))
				if err != nil {
					fmt.Println("Error writing:", err.Error())
					return
				}
				fmt.Println("CONNECT: ERROR")
				return
			}

			fmt.Println("Created session:", id)
			_, err = connection.Write([]byte("CONNECT: OK"))
			if err != nil {
				panic(err)
			}
			data.clientIDs = append(data.clientIDs, id)
		}
	}
}
