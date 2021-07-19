package main

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

type Message struct {
	msg_id       int
	uid          int
	usr_nickname string
	message      string
	time         time.Time
}

type User struct {
	uid      int
	nickname string
}

var broadcast = make(chan Message)

var currMsgId int

const ( //example db data, change to yours
	db_host     = "localhost"
	db_port     = 5432
	db_user     = "postgres"
	db_password = "admin"
	db_name     = "postgres"
)

var db *sql.DB

//global db variable

func main() {

	//db connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		db_host, db_port, db_user, db_password, db_name)

	//open db
	var err error
	db, err = sql.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal("Could not connect to the database: ", err)
	} else {
		log.Printf("Connection to the database is successful.")
	}
	defer db.Close()

	var clientNum int
	clients := make(map[net.Conn]User)

	err = db.QueryRow("SELECT MAX(uid) FROM users").Scan(&clientNum)
	switch {
	case err == sql.ErrNoRows:
		log.Fatalf("no users in db")
	case err != nil:
		log.Fatal(err)
	default:
		log.Printf("Success getting users from DB, max user id = %v", clientNum)
	}

	log.Println("Launching server at port 8000...")
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal("Listening error: ", err)
	}

	defer listener.Close()

	for {

		connection, err := listener.Accept()
		if err != nil {
			log.Fatal("Connection error: ", err)
		}

		go handleRequest(connection, clients, &clientNum)
		go handleMessages(clients)
	}

}

func sendMessage(uid int, usr_nickname string, message string) {
	var msg Message
	msg.msg_id = currMsgId
	msg.uid = uid
	msg.usr_nickname = usr_nickname
	msg.message = message
	broadcast <- msg
}

func handleRequest(connection net.Conn, clients map[net.Conn]User, clientNum *int) {
	defer connection.Close()
	log.Printf("Client #%v connected.", clients[connection])

	//getting user nickname
	clientRequest := bufio.NewReader(connection)
	log.Printf("Waiting for nickname for client#%v", clients[connection].uid)
	msg_string := "Please enter your nickname\n"
	connection.Write([]byte(msg_string))
	clientNickname, _ := clientRequest.ReadString('\n')
	clientNickname = strings.TrimSpace(clientNickname)

	//check if user in db, add into db if user not found
	findIfUserExistsQuery, err := db.Prepare("SELECT * from users where nickname=$1")
	currUser := User{}
	err = findIfUserExistsQuery.QueryRow(clientNickname).Scan(&currUser.uid, &currUser.nickname)
	switch {
	case err == sql.ErrNoRows:
		log.Printf("no user with nickname %s found, creating new user in DB", clientNickname)
		*clientNum++
		//adding user to db
		insertUserQuery := `insert into "users"("uid", "nickname") values($1,$2)`
		_, err = db.Exec(insertUserQuery, *clientNum, clientNickname)
		if err != nil {
			log.Fatal("Couldn't add user to the database: ", err)
		}
		//set current client to user with id clientnum and nickname
		clients[connection] = User{*clientNum, clientNickname}
	case err != nil:
		log.Fatal(err)
	default:
		log.Printf("user with nickname %s found! ", clientNickname)
		clients[connection] = currUser
	}
	defer findIfUserExistsQuery.Close()

	//get all previous messages in chatroom
	err = db.QueryRow("SELECT MAX(msg_id) FROM messages").Scan(&currMsgId)
	switch {
	case err == sql.ErrNoRows:
		log.Fatalf("no messages in db")
	case err != nil:
		log.Fatal(err)
	default:
		log.Printf("success, max msg id = %v", currMsgId)
		getExistingMessagesQuery, err := db.Query("SELECT * from messages")
		if err != nil {
			log.Fatal("Couldn't get existing messages: ", err)
		}
		for getExistingMessagesQuery.Next() {

			currMsg := Message{}
			err := getExistingMessagesQuery.Scan(&currMsg.msg_id, &currMsg.uid, &currMsg.usr_nickname, &currMsg.message, &currMsg.time)
			if err != nil {
				log.Fatal("Error while getting current user: ", err)
			}

			var existing_msg string

			if currMsg.uid != 0 {
				existing_msg = fmt.Sprintf("%v <%v>: %v\n", currMsg.time.Format("15:04"), currMsg.usr_nickname, currMsg.message)
			} else {
				existing_msg = fmt.Sprintf("%v\n", currMsg.message)
			}

			connection.Write([]byte(existing_msg))

		}

		log.Printf("Previous messages sent to user %s.", clients[connection].nickname)

		connection.Write([]byte("\u007F"))

	}

	clientNickname = strings.TrimSpace(clientNickname)
	msg_string = fmt.Sprintf("%v has joined the chat", clientNickname)
	sendMessage(0, clients[connection].nickname, msg_string)

	for {
		clientRequest, err := clientRequest.ReadString('\n')

		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			if clientRequest == "/quit" {
				log.Printf("Client #%v disconnected.",
					clients[connection].uid)
				msg := fmt.Sprintf("%v has left the chat", clientNickname)
				sendMessage(0, clients[connection].nickname, msg)
				delete(clients, connection)
				//*clientNum--
				return
			} else {
				log.Printf("Client #%v sent: %v", clients[connection].uid, clientRequest)
				sendMessage(clients[connection].uid, clients[connection].nickname, clientRequest)
			}
		case io.EOF:
			log.Printf("Client #%v disconnected.",
				clients[connection].uid)
			msg := fmt.Sprintf("%v has left the chat", clientNickname)
			sendMessage(0, clients[connection].nickname, msg)
			delete(clients, connection)
			//*clientNum--
			return
		default:
			log.Printf("Client #%v disconnected.",
				clients[connection].uid)
			msg := fmt.Sprintf("%v has left the chat", clientNickname)
			sendMessage(0, clients[connection].nickname, msg)
			//*clientNum--
			return

		}

	}

}

func handleMessages(clients map[net.Conn]User) {
	for {
		msg := <-broadcast
		time_rn := time.Now()

		var msg_string string

		currMsgId++
		insertMessageQuery := `insert into "messages"("msg_id","uid","usr_nickname", "message", "time") values($1,$2,$3,$4, $5)`
		_, err := db.Exec(insertMessageQuery, currMsgId, msg.uid, msg.usr_nickname, msg.message, time.Now())
		if err != nil {
			log.Fatal("Couldn't add msg to the database: ", err)

		}

		if msg.uid != 0 {
			msg_string = fmt.Sprintf("%v <%v>: %v\n", time_rn.Format("15:04"), msg.usr_nickname, msg.message)
		} else {
			msg_string = fmt.Sprintf("%v\n", msg.message)
		}

		for client := range clients {
			if _, err := client.Write([]byte(msg_string)); err != nil {
				client.Close()
				delete(clients, client)
			}
		}

	}
}
