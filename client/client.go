package client

//go:generate mockgen -destination=mock_client.go -package=client -self_package=github.com/Coderlane/go-minecraft-rcon/client github.com/Coderlane/go-minecraft-rcon/client Client

import (
	"fmt"
	"regexp"

	"github.com/Coderlane/go-minecraft-rcon/rcon"
)

var (
	cmdRegex *regexp.Regexp = regexp.MustCompile(`^[ \w\-\.\;\:\,]+$`)
)

type Client interface {
	Request(cmd string) (string, error)
	Send(cmd string) error
	Close() error
}

// client is a higher level wrapper for RCon connections
type client struct {
	conn *rcon.Conn
}

// Newclient creates a new c that will connect to the RCon server
func NewClient(address, password string) (Client, error) {
	conn, err := rcon.Dial(address, password)
	if err != nil {
		return nil, err
	}
	return &client{
		conn: conn,
	}, nil
}

func validateCommand(cmd string) error {
	if !cmdRegex.MatchString(cmd) {
		return fmt.Errorf("invalid command: %s", cmd)
	}
	return nil
}

// Request sends a request to the server and returns the response
func (c *client) Request(cmd string) (string, error) {
	if err := validateCommand(cmd); err != nil {
		return "", err
	}
	return c.conn.Request(cmd)
}

// Send sends a request to the server and ignores the response
func (c *client) Send(cmd string) error {
	if err := validateCommand(cmd); err != nil {
		return err
	}
	return c.conn.Send(cmd)
}

// Close closes the connection to the server
func (c *client) Close() error {
	return c.conn.Close()
}
