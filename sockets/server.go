// socket-server project main.go
package sockets

import (
	"crypto/rsa"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	clientsFname = "clients.txt"
	serverType   = "tcp"
	serverHost   = "localhost"
)

type ClientData struct {
	clientID         string            // The client's given ID.
	clientData       map[string]string // A mapping of key strings to value strings.
	serverPrivateKey *rsa.PrivateKey
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
func Server(serverPort string) {
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
		buffer, mLen := readClientMessage(connection, clients, id)

		// No message.
		if mLen == 0 {
			continue
		}
		fmt.Printf("User %s: %s\n", id, string(buffer[:mLen]))

		// Last message was PUT [key], current message must be [value].
		if key != "" {
			clients[id].clientData[key] = string(buffer[:mLen])

			if !sendEncrypted(connection, id, "PUT: OK") {
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
				_, err := connection.Write([]byte("CONNECT: ERROR"))
				if err != nil {
					fmt.Println("Error writing:", err.Error())
				}
				return
			}

			privateKey, publicKey := GenerateKeys()
			fmt.Println(KeyToString(publicKey))
			_, err := connection.Write([]byte("CONNECT: " + KeyToString(publicKey)))
			if err != nil {
				fmt.Println("Error writing:", err.Error())
				return
			}

			clients[id] = ClientData{
				clientID:         id,
				clientData:       map[string]string{},
				serverPrivateKey: privateKey,
			}
			defer delete(clients, id)
		// PUT
		case strings.HasPrefix(string(buffer[:mLen]), "PUT "):
			key = string(buffer[4:mLen])
		// GET
		case strings.HasPrefix(string(buffer[:mLen]), "GET "):
			value := []byte(clients[id].clientData[string(buffer[4:mLen])])

			if string(value) == "" {
				if !sendEncrypted(connection, id, "GET: ERROR") {
					return
				}
				continue
			}
			if !sendEncrypted(connection, id, string(value)) {
				return
			}
		// DELETE
		case strings.HasPrefix(string(buffer[:mLen]), "DELETE "):
			_, exists := clients[id].clientData[string(buffer[7:mLen])]
			if !exists {
				if !sendEncrypted(connection, id, "DELETE: ERROR") {
					return
				}
				continue
			}
			delete(clients[id].clientData, string(buffer[7:mLen]))
			if !sendEncrypted(connection, id, "DELETE: OK") {
				return
			}
		// DISCONNECT
		case strings.HasPrefix(string(buffer[:mLen]), "DISCONNECT"):
			if !sendEncrypted(connection, id, "DISCONNECT: OK") {
				return
			}
			return
		// unknown commands
		default:
			if !sendEncrypted(connection, id, "DISCONNECT: UNKNOWN COMMAND") {
				return
			}
			return
		}
	}
}

// sendEncrypted sends the given message along the given connection.
// if an error occurs, sendEncrypted returns false.
func sendEncrypted(connection net.Conn, id, input string) bool {
	publicKey, ok := StringToKey(id)
	if !ok {
		fmt.Println("Error converting string to key")
		return false
	}
	encryptedBytes := Encrypt(publicKey, input)
	_, err := connection.Write(encryptedBytes)
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return false
	}
	if len(id) > 10 {
		fmt.Printf("Send \"%s\" to %s\n", string(encryptedBytes), id[:10])
	} else {
		fmt.Printf("Send \"%s\" to %s\n", string(encryptedBytes), id)
	}
	return true
}

func readClientMessage(
	connection net.Conn,
	clients map[string]ClientData,
	id string,
) ([]byte, int) {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return []byte{}, 0
	}

	if id == "" {
		return buffer, mLen
	}

	privateKey := clients[id].serverPrivateKey
	decryptedBytes, ok := Decrypt(privateKey, buffer[:mLen])
	if !ok {
		fmt.Println("ERROR: failed to decrypt")
		return []byte{}, 0
	}
	return decryptedBytes, len(decryptedBytes)
}
