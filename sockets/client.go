// socket-client project main.go
package sockets

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
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

	var wg sync.WaitGroup
	wg.Add(2)

	// Create a goroutine that will print server responses.
	go readResponses(connection)

	// Create a goroutine that will send user input to server.
	go readUserInput(connection)

	wg.Wait()
	fmt.Println("Program complete")
}

// readResponses will continuously check for server messages and print anything
// it finds. If "DISCONNECT: OK" is received, the function will return.
func readResponses(connection net.Conn) {
	for {
		buffer := make([]byte, 1024)
		mLen, err := connection.Read(buffer)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			return
		}
		fmt.Printf("\n%s\n> ", string(buffer[:mLen]))
	}
}

// readUserInput will continously check for user input and send each line to
// the server. If "DISCONNECT" is input, the function will return.
func readUserInput(connection net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		_, err := connection.Write([]byte(input[:len(input)-2])) // Cut end-line.
		if err != nil {
			fmt.Println("Error writing:", err.Error())
			return
		}
	}
}
