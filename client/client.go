package main

import (
	"bufio"
	"github.com/marcusolsson/tui-go"
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

var posts = []string{}

func main() {
	connection, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		log.Fatal("Connection error: ", err)
	}
	defer connection.Close()



	sidebar := tui.NewVBox(
		tui.NewLabel("CHANNELS"),
		tui.NewLabel("general"),
		tui.NewLabel("random"),
		tui.NewLabel(""),
		tui.NewLabel("DIRECT MESSAGES"),
		tui.NewLabel("slackbot"),
		tui.NewSpacer(),
	)
	sidebar.SetBorder(true)

	history := tui.NewVBox()

	for _, m := range posts {
		history.Append(tui.NewHBox(
			tui.NewLabel(m),
			tui.NewSpacer(),
		))
	}

	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(historyBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)


	root := tui.NewHBox(sidebar, chat)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	input.OnSubmit(func(e *tui.Entry) {
		clientRequest := e.Text()
		clientRequest = strings.TrimSpace(clientRequest)
		if _, err = connection.Write([]byte(clientRequest + "\n")); err != nil {
			log.Printf("failed to send the client request: %v\n", err)
			}
			if clientRequest == "/quit" {
				ui.Quit()
			}
		input.SetText("")
	})

	go getMessages(connection, history, ui)
	ui.SetKeybinding("Esc", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}


}

func getMessages(connection net.Conn, history *tui.Box,ui tui.UI){
	for {
		serverReader := bufio.NewReader(connection)
		serverResponse, err := serverReader.ReadString('\n')
		serverResponse = strings.TrimSpace(serverResponse)
		switch err {
		case nil:
			ui.Update(func() {
				posts = append(posts, serverResponse)
				history.Append(tui.NewHBox(
					tui.NewLabel(serverResponse),
					tui.NewSpacer(),
				))
			})
		case io.EOF:
			log.Println("server closed the connection")
			return
		default:
			log.Printf("server error: %v\n", err)
			return
		}

	}
}
