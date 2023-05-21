package main

import (
	"os"

	"github.com/Rolls71/cosc340-sockets/sockets"
)

// main accepts parameters in the following form:
//   - "client [HOST_NAME] [HOST_PORT]"
//   - "server [HOST_PORT]"
//   - "rsa"
//   - "aes"
func main() {
	switch os.Args[1] {
	case "client":
		sockets.Client(os.Args[2], os.Args[3])
	case "server":
		sockets.Server(os.Args[2])
	case "rsa":
		sockets.TestRSA()
	case "aes":
		sockets.TestAES()
	}
}
