package main

import (
	"os"

	"github.com/Rolls71/cosc340-sockets/sockets"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

func main() {
	switch os.Args[1] {
	case "client":
		sockets.Client(SERVER_HOST, SERVER_PORT)
	case "server":
		sockets.Server(SERVER_HOST, SERVER_PORT)
	}
}
