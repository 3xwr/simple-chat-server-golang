package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	connection, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	defer connection.Close()

	clientReader := bufio.NewReader(os.Stdin)

	go getMessages(connection)
	for {
		clientRequest, err := clientReader.ReadString('\n')

		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			if _, err = connection.Write([]byte(clientRequest + "\n")); err != nil {
				log.Printf("failed to send the client request: %v\n", err)
			}
			if clientRequest == "/quit" {
				os.Exit(0)
			}
		case io.EOF:
			log.Println("client closed the connection")
			return
		default:
			log.Printf("client error: %v\n", err)
			return
		}


	}
}

func getMessages(connection net.Conn){
	for {

	serverReader := bufio.NewReader(connection)
	serverResponse, err := serverReader.ReadString('\n')

	switch err {
	case nil:
		fmt.Println(strings.TrimSpace(serverResponse))
	case io.EOF:
		log.Println("server closed the connection")
		return
	default:
		log.Printf("server error: %v\n", err)
		return
	}

	}
}
