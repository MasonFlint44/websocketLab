package interfaces

import "github.com/gorilla/websocket"

// Client - Defines credentials and connection used to connect to server
type Client interface {
	// GetHandle - Returns handle used to identify user
	GetHandle() string
	// GetPass - Returns password used to authenticate
	GetPass() string
	// GetConn - Returns connection to server
	GetConn() *websocket.Conn
	// SetHandle - Set handle used to identify user
	SetHandle(handle string)
	// SetPass - Set password used to authenticate
	SetPass(pass string)
	// SetConn - Set connection to server
	SetConn(conn *websocket.Conn)
}
