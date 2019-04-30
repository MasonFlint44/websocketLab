package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/masonflint44/websocketLab/pkg/helpers"
	"github.com/masonflint44/websocketLab/pkg/interfaces"
	"github.com/masonflint44/websocketLab/pkg/models"
)

var dialer = websocket.Dialer{}
var inboundMessages = make(chan interfaces.Message)
var outboundMessages = make(chan interfaces.Message)

func main() {
	fmt.Println("Client: Starting...")

	conn, _, err := dialer.Dial("ws://localhost:11631", nil)
	if err != nil {
		fmt.Println("Error: Unable to connect to server")
		return
	}
	defer closeConn(conn)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go receiveMessages(conn, &wg)
	go printMessages()
	go readInput(conn)
	go sendMessages(conn)

	wg.Wait()
}

// sendMessages - Send queued messages to server
func sendMessages(conn *websocket.Conn) {
	for {
		message := <-outboundMessages
		err := conn.WriteJSON(message)
		if err != nil {
			fmt.Println("Error: Unable to send message to server")
		}
	}
}

// readInput - Reads input from stdin to build and queue outbound messages
func readInput(conn *websocket.Conn) {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error: Unable to read input")
			break
		}

		command, body := helpers.SplitOnFirstDelim(' ', line)
		message := models.Message{Command: command, Body: body}

		switch command {
		case "login":
			fallthrough
		case "newuser":
			handle, pass := helpers.SplitOnFirstDelim(' ', message.Body)
			message.Client = models.Client{
				Handle: handle,
				Pass:   pass,
			}
			fallthrough
		case "send":
			fallthrough
		case "logout":
			outboundMessages <- message
			fallthrough
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("- login <handle> <pass> - Log in to server")
			fmt.Println("- newuser <handle> <pass> - Register new user")
			fmt.Println("- send <message> - Send message to clients")
			fmt.Println("- logout - Log out from server")
		default:
			fmt.Println("Type 'help' to get a list available commands")
		}
	}
}

// printMessages - Print queued incoming messages to stdin
func printMessages() {
	for {
		var body string
		message := <-inboundMessages
		client := message.GetClient()
		if client != nil && client.GetHandle() != "" {
			body = client.GetHandle() + ": " + message.GetBody()
		} else {
			body = message.GetBody()
		}
		fmt.Println(body)
	}
}

// receiveMessages - Recieve and queue messages from server
func receiveMessages(conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		var demarshaled struct {
			Command string
			Body    string
			Client  models.Client
		}
		err := conn.ReadJSON(&demarshaled)
		if err != nil {
			fmt.Println("Error: Unable to read message from server")
			break
		}
		message := models.Message{
			Command: demarshaled.Command,
			Body:    demarshaled.Body,
			Client:  demarshaled.Client,
		}
		inboundMessages <- message
	}
}

// closeConn - Close connection to server
func closeConn(conn *websocket.Conn) {
	fmt.Println("Client: Closing connection...")
	conn.Close()
}
