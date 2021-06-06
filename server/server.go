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
	usr_nickname string
	message string
}

type User struct {
	uid int
	nickname string
}

var broadcast = make(chan Message)

func main() {

	clientNum := 0
	clients := make(map[net.Conn]User)

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
		currClient, _  := clients[connection]
		currClient.uid = clientNum
		currClient.nickname = ""
		clients[connection] = currClient

		go handleRequest(connection, clients, &clientNum)
		go handleMessages(clients)
	}

}

func sendMessage(uid int, usr_nickname string, message string) {
	var msg Message
	msg.uid = uid
	msg.usr_nickname = usr_nickname
	msg.message = message
	broadcast <- msg
}

func handleRequest(connection net.Conn, clients map[net.Conn]User, clientNum *int) {
	defer connection.Close()
	log.Printf("Client #%v connected.", clients[connection])

	clientRequest := bufio.NewReader(connection)
	log.Printf("Waiting for nickname for client#%v", clients[connection].uid)
	msg_string := "Please enter your nickname\n"
	connection.Write([]byte(msg_string))

	clientNickname, _ := clientRequest.ReadString('\n')

	currClient, _  := clients[connection]
	currClient.nickname = strings.TrimSpace(clientNickname)
	clients[connection] = currClient

	clientNickname = strings.TrimSpace(clientNickname)
	msg_string = fmt.Sprintf("%v has joined the chat", clientNickname)
	sendMessage(0,clients[connection].nickname,msg_string)


	for {
		clientRequest, err := clientRequest.ReadString('\n')

		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			if clientRequest == "/quit" {
				log.Printf("Client #%v disconnected.",
					clients[connection].uid)
				msg := fmt.Sprintf("%v has left the chat", clientNickname)
				sendMessage(0,clients[connection].nickname,msg)
				delete(clients, connection)
				*clientNum--
				return
			} else {
				log.Printf("Client #%v sent: %v", clients[connection].uid, clientRequest)
				sendMessage(clients[connection].uid,clients[connection].nickname,clientRequest)
			}
		case io.EOF:
			log.Printf("Client #%v disconnected.",
				clients[connection].uid)
			msg := fmt.Sprintf("%v has left the chat", clientNickname)
			sendMessage(0,clients[connection].nickname,msg)
			delete(clients, connection)
			*clientNum--
			return
		default:
			log.Printf("Client #%v disconnected.",
				clients[connection].uid)
			msg := fmt.Sprintf("%v has left the chat", clientNickname)
			sendMessage(0,clients[connection].nickname,msg)
			*clientNum--
			return

		}

	}

}

func handleMessages(clients map[net.Conn]User) {
	for {
		msg := <-broadcast
		var msg_string string
		if msg.uid != 0 {
			msg_string = fmt.Sprintf("%v: %v\n", msg.usr_nickname, msg.message)
		} else {
			msg_string = fmt.Sprintf("%v\n",  msg.message)
		}


		for client := range clients {
			if _,err := client.Write([]byte(msg_string)); err != nil {
				client.Close()
				delete(clients,client)
			}
		}

	}
}

