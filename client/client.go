package client

import (
	"github.com/Coderlane/go-minecraft-rcon/rcon"
)

// Client is a higher level wrapper for RCon connections
type Client struct {
	conn *rcon.Conn
}

// NewClient creates a new client that will connect to the RCon server
func NewClient(address, password string) (*Client, error) {
	conn, err := rcon.Dial(address, password)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn: conn,
	}, nil
}

// Request sends a request to the server and returns the response
func (client *Client) Request(body string) (string, error) {
	return client.conn.Request(body)
}

// Send sends a request to the server and ignores the response
func (client *Client) Send(body string) error {
	return client.conn.Send(body)
}

// Close closes the connection to the server
func (client *Client) Close() error {
	return client.conn.Close()
}
