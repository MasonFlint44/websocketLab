package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/masonflint44/websocketLab/pkg/interfaces"
	"github.com/masonflint44/websocketLab/pkg/models"
)

var upgrader = websocket.Upgrader{}
var clients = make(map[*websocket.Conn]interfaces.Client)

var loginRequests = make(chan request)
var newUserRequests = make(chan request)
var sendRequests = make(chan request)
var logoutRequests = make(chan request)
var helpRequests = make(chan request)
var outboundResponses = make(chan interfaces.Message)

func main() {
	defer func() {
		log.Println("Disconnecting all clients...")
		for conn := range clients {
			disconnect(conn)
		}
	}()

	http.HandleFunc("/", wsHandler)
	go sendMessages()
	go processSendRequests()
	go processLoginRequests()
	go processNewUserRequests()
	go processLogoutRequests()

	log.Printf("Starting server... \n")
	err := http.ListenAndServe(":11631", nil)
	log.Fatal(err)
}

// wsHandler - Upgrade connection to websocket connection
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatal(err)
	}

	client := models.Client{Conn: conn}
	clients[conn] = client
	go receiveMessages(conn)

	client.Handle = "Server"
	outboundResponses <- models.Message{Body: "Welcome to the chat room!", Client: client}

	fmt.Println("Client connected")
}

// receiveMessages - Receives messages for each connected client
func receiveMessages(conn *websocket.Conn) {
	defer disconnect(conn)
	for {
		var demarshaled struct {
			Command string
			Body    string
			Client  models.Client
		}
		err := conn.ReadJSON(&demarshaled)
		if err != nil {
			log.Println("Error: Unable to read message from client")
			log.Println("Disconnecting client...")
			break
		}
		message := models.Message{
			Command: demarshaled.Command,
			Body:    demarshaled.Body,
			Client:  demarshaled.Client,
		}
		request := webRequest{
			Message: message,
			Client:  clients[conn],
		}

		switch command := message.GetCommand(); command {
		case "login":
			loginRequests <- request
		case "newuser":
			newUserRequests <- request
		case "send":
			sendRequests <- request
		case "logout":
			logoutRequests <- request
		case "help":
			helpRequests <- request
		default:
			log.Println("Received unrecognized command -", command, "- from client")
		}
	}
}

// disconnect - Close provided connection
func disconnect(conn *websocket.Conn) {
	conn.Close()
	delete(clients, conn)
}
