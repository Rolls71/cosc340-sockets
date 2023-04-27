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
			clients[id].clientData[key] = string(buffer[:mLen])

			if !sendMessage(connection, id, "PUT: OK") {
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
				if !sendMessage(connection, id, "CONNECT: ERROR") {
					return
				}
				return
			}

			if !sendMessage(connection, id, "CONNECT: OK") {
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
				if !sendMessage(connection, id, "GET: ERROR") {
					return
				}
				continue
			}
			if !sendMessage(connection, id, string(value)) {
				return
			}
		// DELETE
		case strings.HasPrefix(string(buffer[:mLen]), "DELETE "):
			_, exists := clients[id].clientData[string(buffer[7:mLen])]
			if !exists {
				if !sendMessage(connection, id, "DELETE: ERROR") {
					return
				}
				continue
			}
			delete(clients[id].clientData, string(buffer[7:mLen]))
			if !sendMessage(connection, id, "DELETE: OK") {
				return
			}
		// DISCONNECT
		case strings.HasPrefix(string(buffer[:mLen]), "DISCONNECT"):
			if !sendMessage(connection, id, "DISCONNECT: OK") {
				return
			}
			delete(clients, id)
			return
		// unknown commands
		default:
			if !sendMessage(connection, id, "DISCONNECT: UNKNOWN COMMAND") {
				return
			}
			delete(clients, id)
			return
		}
	}
}

// sendMessage sends the given message along the given connection.
// if an error occurs, sendMessage returns false.
func sendMessage(connection net.Conn, id, message string) bool {
	_, err := connection.Write([]byte(message))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return false
	}
	fmt.Printf("Send \"%s\" to %s\n", message, id)
	return true
}
