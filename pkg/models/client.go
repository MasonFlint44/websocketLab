package models

import (
	"github.com/gorilla/websocket"
	"github.com/masonflint44/websocketLab/pkg/interfaces"
)

// Client - Defines credentials and connection used to connect to server
type Client struct {
	// Handle - Handle used to identify user
	Handle string
	// Pass - Password used to authenticate
	Pass string
	// Conn - Connection to server
	Conn *websocket.Conn
}

// GetHandle - Returns handle used to identify user
func (c *Client) GetHandle() string {
	return c.Handle
}

// GetPass - Returns password used to authenticate
func (c *Client) GetPass() string {
	return c.Pass
}

// GetConn - Returns connection to server
func (c *Client) GetConn() *websocket.Conn {
	return c.Conn
}

// SetHandle - Set handle used to identify user
func (c *Client) SetHandle(handle string) {
	c.Handle = handle
}

// SetPass - Set password used to authenticate
func (c *Client) SetPass(pass string) {
	c.Pass = pass
}

// SetConn - Set connection to server
func (c *Client) SetConn(conn *websocket.Conn) {
	c.Conn = conn
}

// CloneClient - Make copy of client
func CloneClient(c interfaces.Client) interfaces.Client {
	return &Client{
		Conn:   c.GetConn(),
		Handle: c.GetHandle(),
		Pass:   c.GetPass(),
	}
}
