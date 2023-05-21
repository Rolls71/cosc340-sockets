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

	clientPrivateKey, clientPublicKey = GenerateKeys()
	serverKey = rsa.PublicKey{}

	// Register session by sending CONNECT message.
	_, err = connection.Write([]byte("CONNECT " + KeyToString(clientPublicKey)))
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
		serverKey, ok = StringToKey(string(buffer[9:mLen]))
		if !ok {
			fmt.Println("ERROR: Received invalid public RSA key")
			os.Exit(1)
		}
		fmt.Println(KeyToString(serverKey))
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
	go readResponses(connection)

	// Create a goroutine that will send user input to server.
	go readUserInput(connection)

	wg.Wait()
}

// readResponses will continuously check for server messages and print anything
// it finds. If "DISCONNECT" is received the program will end.
func readResponses(connection net.Conn) {
	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			os.Exit(1)
		}

		if serverKey != (rsa.PublicKey{}) {
			ok := false
			buffer, ok = Decrypt(clientPrivateKey, buffer[:mLen])
			if !ok {
				panic("ERROR: Failed to decrypt message")
			}
			mLen = len(buffer)
		}

		fmt.Printf("\u001b[0K%s\n> ", string(buffer[:mLen]))
		switch {
		case strings.HasPrefix(string(buffer[:mLen]), "CONNECT"):
			ok := true
			serverKey, ok = StringToKey(string(buffer[9:mLen]))
			if !ok {
				fmt.Println("ERROR: Received invalid public RSA key")
				os.Exit(1)
			}
			fmt.Println(KeyToString(serverKey))
			continue
		case strings.HasPrefix(string(buffer[:mLen]), "PUT"):
		case strings.HasPrefix(string(buffer[:mLen]), "DELETE"):
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

// readUserInput will continously check for user input and send each line to
// the server. If an invalid command is entered, the command will not be sent.
// The function will return if an error sending a message occurs, the
// program will end.
func readUserInput(connection net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		for _, command := range validCommands {
			if !strings.HasPrefix(input, command) && !isPuttingValue {
				continue
			}
			if isPuttingValue {
				sendEncryptedClientMessage(connection, input)
				isPuttingValue = false
				break
			} else if strings.HasPrefix(input, "PUT ") {
				sendEncryptedClientMessage(connection, input)
				isPuttingValue = true
				break
			}
			if strings.HasPrefix(input, "GET ") {
				isGettingValue = true
			}
			sendEncryptedClientMessage(connection, input)
			break
		}
	}
}

func sendEncryptedClientMessage(connection net.Conn, input string) {
	if serverKey == (rsa.PublicKey{}) {
		_, err := connection.Write([]byte(input[:len(input)-endLineChars])) // Cut end-line.
		if err != nil {
			fmt.Println("Error writing:", err.Error())
			os.Exit(1)
		}
		return
	}
	encryptedBytes := Encrypt(serverKey, input[:len(input)-endLineChars])
	_, err := connection.Write(encryptedBytes) // Cut end-line.
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		os.Exit(1)
	}
}
