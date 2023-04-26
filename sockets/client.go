// socket-client project main.go
package sockets

import (
	"fmt"
	"math/rand"
	"net"
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
	connection.Write([]byte("CONNECT " + fmt.Sprintf("%d", id)))
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
		return
	}
	fmt.Println(string(buffer[:mLen]))

	// Send user input to server and handle responses.
	for {
		var response string
		fmt.Scan(&response)
		fmt.Println("User typed: ", response)
		// Write.
	}
}
