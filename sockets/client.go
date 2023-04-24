// socket-client project main.go
package sockets

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

func Client(server_type, server_host, server_port string) {
	rand.Seed(time.Now().UnixNano())
	id := rand.Int()

	//establish connection
	connection, err := net.Dial(server_type, server_host+":"+server_port)
	if err != nil {
		panic(err)
	}
	///send some data
	connection.Write([]byte("CONNECT " + fmt.Sprintf("%d", id)))
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	fmt.Println("> " + string(buffer[:mLen]))
	defer connection.Close()
}
