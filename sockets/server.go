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

type ClientData struct {
	clientID   string            // The client's given ID.
	clientData map[string]string // A mapping of key strings to value strings.
}

// idExists checks if the given client ID string exists in the
// Data.clientIDs slice of strings and returns true if it is found.
func idExists(clientList map[string]ClientData, newClientID string) bool {
	for _, client := range clientList {
		if newClientID == client.clientID {
			return true
		}
	}
	return false
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
	clients := map[string]ClientData{}
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("Client connected")
		go clientSession(connection, clients)
	}
}

// clientSession handles a client connection by repeatedly reading the
// connection buffer and searching for keyword prefixes.
// "CONNECT client_id" will close the connection if the given ID exists. If not,
// add the ID to the Data structure.
func clientSession(connection net.Conn, clients map[string]ClientData) {
	id := ""
	key := ""
	defer connection.Close()
	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}

		// No message.
		if mLen == 0 {
			continue
		}
		fmt.Printf("User %s: %s\n", id, string(buffer[:mLen]))

		// Last message was PUT [key], current message must be [value].
		if key != "" {
			fmt.Printf("Storing \"%s\": \"%s\" under user %s\n", key, string(buffer[:mLen]), id)
			clients[id].clientData[key] = string(buffer[:mLen])

			_, err = connection.Write([]byte("PUT: OK"))
			if err != nil {
				fmt.Println("Error writing:", err.Error())
				return
			}

			// Clear key for next PUT.
			key = ""
			continue
		}

		switch {
		// CONNECT
		case strings.HasPrefix(string(buffer[:mLen]), "CONNECT "):
			id = string(buffer[8:mLen])
			if idExists(clients, id) {
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
				fmt.Println("Error writing:", err.Error())
				return
			}
			clients[id] = ClientData{
				clientID:   id,
				clientData: map[string]string{},
			}
		// PUT
		case strings.HasPrefix(string(buffer[:mLen]), "PUT "):
			key = string(buffer[4:mLen])
		// GET
		case strings.HasPrefix(string(buffer[:mLen]), "GET "):
			value := []byte(clients[id].clientData[string(buffer[4:mLen])])

			if string(value) == "" {
				_, err = connection.Write([]byte("GET: ERROR"))
				if err != nil {
					fmt.Println("Error writing:", err.Error())
					return
				}
				continue
			}

			_, err = connection.Write(value)
			if err != nil {
				fmt.Println("Error writing:", err.Error())
				return
			}
			fmt.Printf("Sent value \"%s\" to user %s\n", value, id)
		}
	}
}
