package main

import (
	"os"

	"github.com/Rolls71/cosc340-sockets/sockets"
)

//	SERVER_HOST = "localhost"
//	SERVER_PORT = "9988"
//	SERVER_TYPE = "tcp"

func main() {
	switch os.Args[1] {
	case "client":
		sockets.Client(os.Args[2], os.Args[3])
	case "server":
		sockets.Server(os.Args[2])
	}
}
