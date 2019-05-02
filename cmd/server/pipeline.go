package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"

	"github.com/masonflint44/websocketLab/pkg/helpers"
	"github.com/masonflint44/websocketLab/pkg/interfaces"
	"github.com/masonflint44/websocketLab/pkg/models"
)

// TODO: update documentation
// TODO: test updated pipeline

func errorPipe(err error, processors ...func(error) error) error {
	for _, processor := range processors {
		if err == nil {
			return nil
		}
		err = processor(err)
	}
	return err
}

// messagePipe - Takes a message and a list of message processors.
// The processors are chained together - the input of the previous is passed in as the parameter to the next.
// If a processor returns nil, the pipeline halts.
// If all processors compelete the function returns true, otherwise false.
func messagePipe(message interfaces.Message, err error, processors ...func(interfaces.Message) (interfaces.Message, error)) (nextMessage interfaces.Message, nextErr error) {
	nextMessage, nextErr = message, err
	for _, processor := range processors {
		if nextErr != nil {
			return
		}
		nextMessage, nextErr = processor(nextMessage)
	}
	return
}

// clientPipe - Takes a client and a list of client processors.
// The processors are chained together - the input of the previous is passed in as the parameter to the next.
// If a processor returns nil, the pipeline halts.
// If all processors compelete the function returns true, otherwise false.
func clientPipe(client interfaces.Client, err error, processors ...func(interfaces.Client) (interfaces.Client, error)) (nextClient interfaces.Client, nextErr error) {
	nextClient, nextErr = client, err
	for _, processor := range processors {
		if nextErr != nil {
			return
		}
		nextClient, nextErr = processor(nextClient)
	}
	return
}

func printError(err error) error {
	log.Println(err)
	return err
}

func catchError(err error) error {
	return nil
}

func toClientErrorHandler(handler func(error) error) func(interfaces.Client, error) (interfaces.Client, error) {
	return func(client interfaces.Client, err error) (interfaces.Client, error) {
		err = handler(err)
		return client, err
	}
}

func hasMessage(message interfaces.Message) (interfaces.Message, error) {
	if message == nil {
		return message, errors.New("Message is nil")
	}
	return message, nil
}

func setClient(client interfaces.Client) func(message interfaces.Message) (interfaces.Message, error) {
	return func(message interfaces.Message) (interfaces.Message, error) {
		message.SetClient(client)
		return message, nil
	}
}

// sendMessage - Send message to client
func sendMessage(message interfaces.Message) (interfaces.Message, error) {
	err := message.GetClient().GetConn().WriteJSON(message)
	return message, err
}

func clientProcessorToErrorHandler(processor func(interfaces.Client) (interfaces.Client, error)) func(interfaces.Client, error) (interfaces.Client, error) {
	return func(client interfaces.Client, err error) (interfaces.Client, error) {
		nextClient, err := processor(client)
		return nextClient, err
	}
}

func onClientError(processor func(interfaces.Client) (interfaces.Client, error), handlers ...func(interfaces.Client, error) (interfaces.Client, error)) func(interfaces.Client) (interfaces.Client, error) {
	return func(client interfaces.Client) (interfaces.Client, error) {
		nextClient, err := processor(client)
		if err == nil {
			return nextClient, err
		}
		for _, handler := range handlers {
			nextClient, nextErr := handler(nextClient, err)
			if nextErr != nil {
				return nextClient, nextErr
			}
		}
		return nextClient, err
	}
}

func catchClientError(processor func(interfaces.Client) (interfaces.Client, error), handlers ...func(interfaces.Client, error) (interfaces.Client, error)) func(interfaces.Client) (interfaces.Client, error) {
	return func(client interfaces.Client) (interfaces.Client, error) {
		nextClient, err := processor(client)
		if err == nil {
			return nextClient, nil
		}
		innerClient, innerErr := nextClient, err
		for _, handler := range handlers {
			innerClient, innerErr = handler(innerClient, innerErr)
			if innerErr == nil {
				return nextClient, nil
			}
		}
		return nextClient, nil
	}
}

func queueMessageToClient(message interfaces.Message) func(interfaces.Client) (interfaces.Client, error) {
	return func(client interfaces.Client) (interfaces.Client, error) {
		message.SetClient(client)
		_, err := messagePipe(message, nil, queueMessage)
		return client, err
	}
}

func queueCustomMessageToClient(handle string, body string) func(interfaces.Client) (interfaces.Client, error) {
	return func(client interfaces.Client) (interfaces.Client, error) {
		client.SetHandle(handle)
		_, err := messagePipe(&models.Message{Body: body, Client: client}, nil, queueMessage)
		return client, err
	}
}

func sendMessageToClient(message interfaces.Message) func(interfaces.Client) (interfaces.Client, error) {
	return func(client interfaces.Client) (interfaces.Client, error) {
		_, err := messagePipe(message, nil,
			hasMessage,
			setClient(client),
			sendMessage,
		)
		return client, err
	}
}

func hasClient(client interfaces.Client) (interfaces.Client, error) {
	if client == nil {
		return client, errors.New("Client is nil")
	}
	return client, nil
}

// hasConn - Evaluates if client has a connection
func hasConn(client interfaces.Client) (interfaces.Client, error) {
	if client.GetConn() == nil {
		return client, errors.New("Client does not have connection")
	}
	return client, nil
}

// hasAuth - Evaluates if client is authenticated
func hasAuth(client interfaces.Client) (interfaces.Client, error) {
	if client.GetHandle() == "" {
		return client, errors.New("Client has not been authenticated")
	}
	return client, nil
}

// validHandle - Evaluates if client has a valid handle
func validHandle(client interfaces.Client) (interfaces.Client, error) {
	if len(client.GetHandle()) > 32 {
		return client, errors.New("Handle must be less than 32 characters")
	}
	return client, nil
}

// validPass - Ensures client has a valid password
func validPass(client interfaces.Client) (interfaces.Client, error) {
	length := len(client.GetPass())
	if length < 4 || length > 8 {
		return client, errors.New("Pass must be between 4 and 8 characters")
	}
	return client, nil
}

func setHandle(source interfaces.Client) func(client interfaces.Client) (interfaces.Client, error) {
	return func(target interfaces.Client) (interfaces.Client, error) {
		target.SetHandle(source.GetHandle())
		return target, nil
	}
}

func setConn(source interfaces.Client) func(client interfaces.Client) (interfaces.Client, error) {
	return func(target interfaces.Client) (interfaces.Client, error) {
		target.SetConn(source.GetConn())
		return target, nil
	}
}

// logout - Log out provided client
func logout(client interfaces.Client) (interfaces.Client, error) {
	clients[client.GetConn()] = &models.Client{Conn: client.GetConn()}
	return client, nil
}

// queueMessage - Push message to outbound queue
func queueMessage(message interfaces.Message) (interfaces.Message, error) {
	outboundResponses <- message
	return message, nil
}

// register - Register client as a new user
func register(client interfaces.Client) (interfaces.Client, error) {
	file, err := os.OpenFile("users.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return client, err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	defer writer.Flush()
	_, err = writer.WriteString("\n" + client.GetHandle() + "," + client.GetPass())
	if err != nil {
		return client, err
	}
	return client, err
}

// uniqueHandle - Ensures client has a unique handle
func uniqueHandle(client interfaces.Client) (interfaces.Client, error) {
	file, err := os.Open("users.txt")
	if err != nil {
		return client, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return client, err
		}
		handle, _ := helpers.SplitOnFirstDelim(',', line)
		if client.GetHandle() == handle {
			return client, errors.New("Handle is not unique")
		}
		if err == io.EOF {
			break
		}
	}
	return client, err
}

// authorize - Log in existing user
func authorize(messageClient interfaces.Client) func(interfaces.Client) (interfaces.Client, error) {
	return func(requestClient interfaces.Client) (interfaces.Client, error) {
		file, err := os.Open("users.txt")
		if err != nil {
			// Unable to open login credentials source
			return requestClient, err
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				// Unable to read from login credentials source
				return requestClient, err
			}
			handle, pass := helpers.SplitOnFirstDelim(',', line)
			if messageClient.GetHandle() == handle && messageClient.GetPass() == pass {
				clients[requestClient.GetConn()] = &models.Client{
					Conn:   requestClient.GetConn(),
					Handle: handle,
				}
				return requestClient, nil
			}
			if err == io.EOF {
				break
			}
		}
		return requestClient, errors.New("Login credentials not in file")
	}
}

// forEachClient - Perform provided function for each client on server
func forEachClient(err error, processors ...func(interfaces.Client) (interfaces.Client, error)) error {
	for _, client := range clients {
		if err != nil {
			return err
		}
		_, err = clientPipe(client, err, processors...)
	}
	return err
}
