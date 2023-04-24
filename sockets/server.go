// socket-server project main.go
package sockets

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	clientsFname = "clients.txt"
)

func Server(server_type, server_host, server_port string) {
	fmt.Println("Server Running...")
	server, err := net.Listen(server_type, server_host+":"+server_port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Listening on " + server_host + ":" + server_port)
	fmt.Println("Waiting for client...")
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("client connected")
		go processClient(connection)
	}
}
func processClient(connection net.Conn) {
	buffer := make([]byte, 1024)
	mLen, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	if strings.HasPrefix(string(buffer[:mLen]), "CONNECT ") {
		sessionID := string(buffer[8:mLen])
		if isIdTaken(sessionID) {
			_, err = connection.Write([]byte("CONNECT: ERROR"))
			if err != nil {
				panic(err)
			}
			fmt.Println("CONNECT: ERROR")
			connection.Close()
			return
		}

		fmt.Println("Created session:", string(buffer[8:mLen]))
		_, err = connection.Write([]byte("CONNECT: OK"))
		if err != nil {
			panic(err)
		}
		storeID(sessionID)

		/*
			f, err := os.Create(string(buffer[8:mLen]) + ".txt")
			if err != nil {
				log.Fatal(err)
			}
			defer deleteFile(f.Name())*/
	}

}

func isIdTaken(id string) bool {
	f, err := os.Open(clientsFname)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() == id {
			return true
		}
	}

	err = scanner.Err()
	if err != nil {
		panic(err)
	}
	return false
}

func storeID(id string) {
	f, err := os.OpenFile(clientsFname, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, "\n"+id)
	if err != nil {
		panic(err)
	}
}
