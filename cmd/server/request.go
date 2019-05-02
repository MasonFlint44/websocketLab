package main

import (
	"github.com/masonflint44/websocketLab/pkg/interfaces"
)

// request - Defines request for server processing
type request interface {
	// GetMessage - Used to get message sent by the client
	GetMessage() interfaces.Message
	// GetClient - Used to get client who made the request
	GetClient() interfaces.Client
}

// serverRequest - Implementation of request for server processing
type serverRequest struct {
	// Message - Message sent by client
	Message interfaces.Message
	// Client - Client who made the request
	Client interfaces.Client
}

// GetMessage - Used to get message sent by the client
func (r serverRequest) GetMessage() interfaces.Message {
	return r.Message
}

// GetClient - Used to get client who made the request
func (r serverRequest) GetClient() interfaces.Client {
	return r.Client
}
