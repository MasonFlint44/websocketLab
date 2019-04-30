package main

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/gorilla/websocket"
	"github.com/masonflint44/websocketLab/pkg/helpers"
	"github.com/masonflint44/websocketLab/pkg/interfaces"
	"github.com/masonflint44/websocketLab/pkg/models"
)

// messagePipe - Takes a message and a list of message processors.
// The processors are chained together - the input of the previous is passed in as the parameter to the next.
// If a processor returns nil, the pipeline halts.
// If all processors compelete the function returns true, otherwise false.
func messagePipe(m interfaces.Message, processors ...func(interfaces.Message) interfaces.Message) func() bool {
	return func() bool {
		processing := m
		if processing == nil {
			return false
		}
		for _, processor := range processors {
			processing = processor(processing)
			if processing == nil {
				return false
			}
		}
		return true
	}
}

// clientPipe - Takes a client and a list of client processors.
// The processors are chained together - the input of the previous is passed in as the parameter to the next.
// If a processor returns nil, the pipeline halts.
// If all processors compelete the function returns true, otherwise false.
func clientPipe(c interfaces.Client, processors ...func(interfaces.Client) interfaces.Client) func() bool {
	return func() bool {
		processing := c
		if processing == nil {
			return false
		}
		for _, processor := range processors {
			processing = processor(processing)
			if processing == nil {
				return false
			}
		}
		return true
	}
}

// pipe - Used to connect different types of pipelines.
// If a pipeline returns false, the pipeline halts.
// If all pipelines complete the function returns true, otherwise false.
func pipe(pipelines ...func() bool) func() bool {
	return func() bool {
		for _, pipeline := range pipelines {
			if pipeline() == false {
				return false
			}
		}
		return true
	}
}

// failPipe - Used to execute a pipeline if the provided condition is false.
// Does not execute pipelines if the provided condition is true.
// Returns true if the provided condition is true, otherwise returns the result of the executed pipelines.
func failPipe(condition bool, pipelines ...func() bool) func() bool {
	if condition == true {
		return func() bool {
			return true
		}
	}
	return pipe(pipelines...)
}

// closePipe - Used to halt a pipeline
func closePipe() func() bool {
	return func() bool {
		return false
	}
}

// tap - Executes the provided function as part of a pipeline
func tap(do func()) func() bool {
	return func() bool {
		do()
		return true
	}
}

// hasConn - Evaluates if client has a connection
func hasConn(c interfaces.Client) interfaces.Client {
	if c.GetConn() == nil {
		return nil
	}
	return c
}

// hasAuth - Evaluates if client is authenticated
func hasAuth(c interfaces.Client) interfaces.Client {
	if c.GetHandle() == "" {
		return nil
	}
	return c
}

// validHandle - Evaluates if client has a valid handle
func validHandle(c interfaces.Client) interfaces.Client {
	if len(c.GetHandle()) > 32 {
		return nil
	}
	return c
}

// uniqueHandle - Ensures client has a unique handle
func uniqueHandle(c interfaces.Client) interfaces.Client {
	file, err := os.Open("users.txt")
	if err != nil {
		log.Println("Error: Unable to open login credentials source")
		return nil
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Println("Error: Unable to read from login credentials source")
			break
		}
		handle, _ := helpers.SplitOnFirstDelim(',', line)
		if c.GetHandle() == handle {
			return nil
		}
		if err == io.EOF {
			break
		}
	}
	return c
}

// register - Register client as a new user
func register(c interfaces.Client) interfaces.Client {
	file, err := os.OpenFile("users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Println("Error: Unable to open login credentials source")
		return nil
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	_, err = writer.WriteString("\n" + c.GetHandle() + "," + c.GetPass())
	if err != nil {
		log.Println("Error: Unable to write to login credentials source")
		return nil
	}
	return c
}

// validPass - Ensures client has a valid password
func validPass(c interfaces.Client) interfaces.Client {
	length := len(c.GetPass())
	if length < 4 || length > 8 {
		return nil
	}
	return c
}

// logout - Log out provided client
func logout(c interfaces.Client) interfaces.Client {
	client := models.Client{Conn: c.GetConn()}
	clients[c.GetConn()] = client
	return client
}

// queueMessage - Push message to outbound queue
func queueMessage(m interfaces.Message) interfaces.Message {
	outboundResponses <- models.Message{Body: m.GetBody(), Client: m.GetClient()}
	return m
}

// forEachClient - Perform provided function for each client on server
func forEachClient(do func(interfaces.Message) interfaces.Message) []func(interfaces.Message) interfaces.Message {
	dos := make([]func(interfaces.Message) interfaces.Message, 0)
	for conn := range clients {
		dos = append(
			dos,
			func(m interfaces.Message) interfaces.Message {
				client := models.Client{
					Conn: conn,
				}
				if m.GetClient() != nil {
					client.Handle = m.GetClient().GetHandle()
				}
				return do(models.Message{
					Body:    m.GetBody(),
					Command: m.GetCommand(),
					Client:  client,
				})
			})
	}
	return dos
}

// authorize - Log in existing user
func authorize(conn *websocket.Conn) func(interfaces.Client) interfaces.Client {
	return func(messageClient interfaces.Client) interfaces.Client {
		file, err := os.Open("users.txt")
		if err != nil {
			log.Println("Error: Unable to open login credentials source")
			return nil
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				log.Println("Error: Unable to read from login credentials source")
				break
			}
			handle, pass := helpers.SplitOnFirstDelim(',', line)
			if messageClient.GetHandle() == handle && messageClient.GetPass() == pass {
				clients[conn] = models.Client{
					Conn:   conn,
					Handle: handle,
				}
				return messageClient
			}
			if err == io.EOF {
				break
			}
		}
		return nil
	}
}

// sendResponse - Send message to client
func sendResponse(m interfaces.Message) interfaces.Message {
	err := m.GetClient().GetConn().WriteJSON(m)
	if err != nil {
		log.Println("Error: Unable to send message to client")
		return nil
	}
	return m
}
