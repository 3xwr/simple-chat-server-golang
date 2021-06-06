package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type Message struct {
	uid int
	message string
}

var broadcast = make(chan Message)

func main() {

	clientNum := 0
	clients := make(map[net.Conn]int)

	log.Println("Launching server at port 8000...")
	listener,err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal("Listening error: ", err)
	}

	defer listener.Close()

	for {

		connection,err := listener.Accept()
		if err != nil {
			log.Fatal("Connection error: ", err)
		}

		clientNum = clientNum+1
		clients[connection] = clientNum
		go handleRequest(connection, clients, &clientNum)
		go handleMessages(clients)
	}

}

func handleRequest(connection net.Conn, clients map[net.Conn]int, clientNum *int) {
	defer connection.Close()
	log.Printf("Client #%v connected.", clients[connection])

	clientRequest := bufio.NewReader(connection)

	for {
		clientRequest, err := clientRequest.ReadString('\n')

		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			log.Printf("Client #%v sent: %v", clients[connection], clientRequest)

			var msg Message
			msg.uid=clients[connection]
			msg.message = clientRequest
			broadcast <- msg

		case io.EOF:
			log.Printf("Client #%v disconnected.",
				clients[connection])
			delete(clients, connection)
			*clientNum--
			return
		default:
			log.Printf("Client #%v disconnected.",
				clients[connection])
			*clientNum--
			return

		}

	}

}

func handleMessages(clients map[net.Conn]int) {
	for {
		msg := <-broadcast
		msg_string := fmt.Sprintf("Client #%v sent: %v\n", msg.uid, msg.message)

		for client := range clients {
			if _,err := client.Write([]byte(msg_string)); err != nil {
				client.Close()
				delete(clients,client)
			}
		}

	}
}
