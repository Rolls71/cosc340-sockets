// socket-client project main.go
package sockets

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func Client(serverHost, serverPort string) {
	// Set seed to current time and generate random client ID.
	rand.Seed(time.Now().UnixNano())
	id := rand.Int()

	// Connect to server and close connection upon return.
	connection, err := net.Dial(serverType, serverHost+":"+serverPort)
	if err != nil {
		panic(err)
	}
	defer connection.Close()

	// Register session by sending CONNECT message.
	_, err = connection.Write([]byte("CONNECT " + fmt.Sprintf("%d", id)))
	if err != nil {
		fmt.Println("Error writing:", err.Error())
		return
	}

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
		fmt.Printf("\n%s\n> ", string(buffer[:mLen]))
		if strings.HasPrefix(string(buffer[:mLen]), "DISCONNECT") {
			os.Exit(0)
		}
	}
}

// readUserInput will continously check for user input and send each line to
// the server. The function will return if an error sending a message occurs,
// the program will end.
func readUserInput(connection net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		_, err := connection.Write([]byte(input[:len(input)-2])) // Cut end-line.
		if err != nil {
			fmt.Println("Error writing:", err.Error())
			os.Exit(1)
		}
	}
}
