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
	serverPrivateKey *rsa.PrivateKey   // The keys used for conversation with this client.
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
		// If the last client message was PUT [key], the current message must
		// be [value]. Skip validation
		if key != "" {
			buffer := make([]byte, 1024)
			mLen, err := connection.Read(buffer)
			if err != nil {
				fmt.Println("Error reading message from "+id+": ", err.Error())
				return
			}
			clients[id].clientData[key] = string(buffer[:mLen])

			if !sendServerMessage(connection, id, "PUT: OK") {
				return
			}

			// Clear key for next PUT command.
			key = ""
			continue
		}

		buffer, mLen, ok := readClientMessage(connection, clients, id)
		if !ok {
			return
		}

		// No message.
		if mLen == 0 {
			continue
		}
		if len(id) > 10 {
			fmt.Printf("User %s...: %s\n", id[:10], string(buffer[:mLen]))
		} else {
			fmt.Printf("User %s: %s\n", id, string(buffer[:mLen]))
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

			privateKey, publicKey := GenerateRSAKeys()
			_, err := connection.Write([]byte("CONNECT: " + RSAKeyToString(publicKey)))
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
				if !sendServerMessage(connection, id, "GET: ERROR") {
					return
				}
				continue
			}
			if !sendServerMessage(connection, id, string(value)) {
				return
			}
		// DELETE
		case strings.HasPrefix(string(buffer[:mLen]), "DELETE "):
			_, exists := clients[id].clientData[string(buffer[7:mLen])]
			if !exists {
				if !sendServerMessage(connection, id, "DELETE: ERROR") {
					return
				}
				continue
			}
			delete(clients[id].clientData, string(buffer[7:mLen]))
			if !sendServerMessage(connection, id, "DELETE: OK") {
				return
			}
		// DISCONNECT
		case strings.HasPrefix(string(buffer[:mLen]), "DISCONNECT"):
			if !sendServerMessage(connection, id, "DISCONNECT: OK") {
				return
			}
			return
		// Unknown commands.
		default:
			if !sendServerMessage(connection, id, "DISCONNECT: UNKNOWN COMMAND") {
				return
			}
			return
		}
	}
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

// sendServerMessage applies RSA encryption to the given input, and sends it
// along the given conection.
//
// Returns false if an error occurs.
func sendServerMessage(connection net.Conn, id, input string) bool {
	publicKey, ok := StringToRSAKey(id)
	if !ok {
		fmt.Println("Error converting string to key")
		return false
	}
	encryptedBytes, ok := EncryptRSA(publicKey, input)
	if !ok {
		fmt.Println("Error encrypting message")
		return false
	}
	_, err := connection.Write(encryptedBytes)
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return false
	}
	if len(id) > 10 {
		fmt.Printf("Send \"%s\" to %s...\n", string(input), id[:10])
	} else {
		fmt.Printf("Send \"%s\" to %s\n", string(input), id)
	}
	return true
}

// readClientMessage reads from the given connection and RSA decrypts the
// message if keys have been exchanged. Otherwise it returns the message as is.
//
// Returns a byte array of the clients message and a boolean indicating success.
func readClientMessage(
	connection net.Conn,
	clients map[string]ClientData,
	id string,
) ([]byte, int, bool) {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading message from "+id+": ", err.Error())
		return []byte{}, 0, false
	}

	if id == "" {
		return buffer, mLen, true
	}

	privateKey := clients[id].serverPrivateKey
	decryptedBytes, ok := DecryptRSA(privateKey, buffer[:mLen])
	if !ok {
		fmt.Println("ERROR: failed to decrypt")
		return []byte{}, 0, false
	}
	return decryptedBytes, len(decryptedBytes), true
}
