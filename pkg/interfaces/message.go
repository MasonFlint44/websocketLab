package interfaces

// Message - Defines a message that includes the following:
type Message interface {
	// GetCommand - Used to allow the processor to determine how to interpret the message
	GetCommand() string
	// GetBody - Body of the message
	GetBody() string
	// GetClient - Returns information about client sending the message
	GetClient() Client
	// SetCommand - Used to allow the processor to determine how to interpret the message
	SetCommand(command string)
	// SetBody - Set body of the message
	SetBody(body string)
	// SetClient - Set information about client sending the message
	SetClient(client Client)
}
