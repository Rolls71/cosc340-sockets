// socket-client project main.go
package sockets

import (
	"bufio"
	"crypto/rsa"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
)

var validCommands = [4]string{"PUT ", "GET ", "DELETE ", "DISCONNECT"}
var isPuttingValue = false
var isGettingValue = false
var endLineChars = 2
var clientPrivateKey *rsa.PrivateKey
var clientPublicKey rsa.PublicKey
var serverKey rsa.PublicKey
var aesKey []byte

// Client attempts to establish a socket connection to a TCP server with the
// given host name and port. After using net.Dial, Client will generate RSA keys
// and an AES key for secure communication and data storage. Client will then
// run two goroutines that continuously read user input and server input.
func Client(serverHost, serverPort string) {
	if runtime.GOOS == "windows" {
		endLineChars = 2
	} else {
		endLineChars = 1
	}

	// Connect to server and close connection upon return.
	connection, err := net.Dial(serverType, serverHost+":"+serverPort)
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	clientPrivateKey, clientPublicKey = GenerateRSAKeys()
	serverKey = rsa.PublicKey{}
	aesKey = GenerateAESKey()

	// Register session by sending CONNECT message.
	_, err = connection.Write([]byte("CONNECT " + RSAKeyToString(clientPublicKey)))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return
	}

	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		os.Exit(1)
	}
	if string(buffer[:mLen]) == "CONNECT: ERROR" {
		fmt.Println("Error: session ID is already taken.")
		os.Exit(1)
	}
	if strings.HasPrefix(string(buffer[:mLen]), "CONNECT") {
		ok := true
		serverKey, ok = StringToRSAKey(string(buffer[9:mLen]))
		if !ok {
			fmt.Println("ERROR: Received invalid public RSA key")
			os.Exit(1)
		}
		fmt.Println(RSAKeyToString(serverKey))
	}

	fmt.Println(`
KEY-VALUE STORE CLIENT
Connects to the given server and manipulates data with the following commands:
* PUT [key] - Allows the client to store a key, the following message will be stored as the associated value. 
The server responds \"PUT: OK\" or \"PUT: ERROR\", depending on whether the operation is successful.
* GET [key] - Allows the client to retrieve the value associated with a given key, if such a value exists. 
The server responds either with the associated value or a \"GET: ERROR\" message.
* DELETE [key] - Allows the client to delete a key and its associated value. The server responds \"DELETE: OK\" 
or \"DELETE: ERROR\", depending on whether the operation is successful.
* DISCONNECT - The server will remove all values stored by the client from its system and respond \"DISCONNECT: OK\". 
After receiving a \"DISCONNECT: OK\" message, the client exits.
After sending any other than these commands, the server and client will disconnect.`)

	// Wait for goroutines to return before ending program.
	var wg sync.WaitGroup
	wg.Add(2)

	// Create a goroutine that will print server responses.
	go readServerMessages(connection)

	// Create a goroutine that will send user input to server.
	go readUserInputs(connection)

	wg.Wait()
}

// readServerMessages will continuously check for server messages and RSA
// decrypt them. If "DISCONNECT" or an unrecognised command is received, the
// client will disconnect.
func readServerMessages(connection net.Conn) {
	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			os.Exit(1)
		}

		if serverKey != (rsa.PublicKey{}) {
			ok := false
			buffer, ok = DecryptRSA(clientPrivateKey, buffer[:mLen])
			if !ok {
				panic("ERROR: Failed to decrypt message")
			}
			mLen = len(buffer)
		}

		if mLen == 0 {
			continue
		}

		if isGettingValue && string(buffer[:mLen]) != "GET: ERROR" {
			plaintext, ok := DecryptAES(aesKey, buffer[:mLen])
			if !ok {
				fmt.Println("Error during AES decryption")
				os.Exit(1)
			}
			fmt.Printf("\u001b[0K%s\n> ", plaintext)
		} else {
			fmt.Printf("\u001b[0K%s\n> ", string(buffer[:mLen]))
		}
		switch {
		case strings.HasPrefix(string(buffer[:mLen]), "CONNECT: "):
			ok := true
			serverKey, ok = StringToRSAKey(string(buffer[9:mLen]))
			if !ok {
				fmt.Println("ERROR: Received invalid public RSA key")
				os.Exit(1)
			}
			fmt.Println(RSAKeyToString(serverKey))
			continue
		case strings.HasPrefix(string(buffer[:mLen]), "PUT: "):
		case strings.HasPrefix(string(buffer[:mLen]), "DELETE: "):
			continue
		default:
			if isGettingValue {
				isGettingValue = false
				continue
			} else {
				os.Exit(0)
			}
		}
	}
}

// readUserInputs will continously check for user input and send each line to
// the server. If an invalid command is entered, the command will not be sent.
// The client will disconnect if an error occurs
func readUserInputs(connection net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		for _, command := range validCommands {
			if !strings.HasPrefix(input, command) && !isPuttingValue {
				continue
			}
			if isPuttingValue {
				ciphertext, ok := EncryptAES(aesKey, input[:len(input)-endLineChars])
				if !ok {
					fmt.Println("Error during AES encryption")
					os.Exit(1)
				}
				_, err := connection.Write([]byte(ciphertext)) // Cut end-line.
				if err != nil {
					fmt.Println("Error writing:", err.Error())
					os.Exit(1)
				}
				isPuttingValue = false
				break
			} else if strings.HasPrefix(input, "PUT ") {
				sendClientMessage(connection, input)
				isPuttingValue = true
				break
			}
			if strings.HasPrefix(input, "GET ") {
				isGettingValue = true
			}
			sendClientMessage(connection, input)
			break
		}
	}
}

// sendClientMessage will RSA encrypt the given message and send it along the
// given connection. If public keys have not yet been exchanged, the message
// will not be encrypted. Disconnects the client if an error occurs.
func sendClientMessage(connection net.Conn, input string) {
	if serverKey == (rsa.PublicKey{}) {
		_, err := connection.Write([]byte(input[:len(input)-endLineChars])) // Cut end-line.
		if err != nil {
			fmt.Println("Error writing:", err.Error())
			os.Exit(1)
		}
		return
	}
	encryptedBytes, ok := EncryptRSA(serverKey, input[:len(input)-endLineChars])
	if !ok {
		fmt.Println("Error encrypting message")
		os.Exit(1)
	}
	_, err := connection.Write(encryptedBytes) // Cut end-line.
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		os.Exit(1)
	}
}
