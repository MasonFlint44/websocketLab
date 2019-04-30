package models

import "github.com/masonflint44/websocketLab/pkg/interfaces"

// Message - Defines a message that includes the following:
type Message struct {
	// Command - Used to allow the processor to determine how to interpret the message
	Command string
	// Body - Body of the message
	Body string
	// Client - Information about client sending the message
	Client interfaces.Client
}

// GetCommand - Used to allow the processor to determine how to interpret the message
func (m Message) GetCommand() string {
	return m.Command
}

// GetBody - Body of the message
func (m Message) GetBody() string {
	return m.Body
}

// GetClient - Returns information about client sending the message
func (m Message) GetClient() interfaces.Client {
	return m.Client
}
