package rcon

import (
	"fmt"
	"log"
	"math"
	"net"
	"strings"
	"time"
)

// ResponseCallback is used in server handlers,
type ResponseCallback func(resp string) error

// Handler handles rcon commands. |cmd\ represents the full command and
// |cb| can be called with the response to the command if there is one.
//
// NOTE: To best emulate a real rcon server, do not handle this on a new
// goroutine.
type Handler interface {
	ServeRCon(cb ResponseCallback, cmd string) error
}

// Server is a minimal RCon protocol server that handles multiple connections
// simultaneously, but only one request per connection at a time.
type Server struct {
	password string
	listener net.Listener
	handlers map[string]Handler
	respChan chan error
}

// Listen on address for new connections. Only accept them if password is
// provided.
func Listen(address, password string) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	srv := &Server{
		password: password,
		listener: listener,
		handlers: make(map[string]Handler),
		respChan: make(chan error),
	}
	// Start accepting connections
	go func() {
		for {
			conn, err := srv.listener.Accept()
			if err != nil {
				return
			}
			conn.SetDeadline(time.Time{})
			go srv.handleConnection(conn)
		}
	}()
	return srv, nil
}

// Addr returns the address the server is listening on
func (srv *Server) Addr() net.Addr {
	return srv.listener.Addr()
}

// Close closes the connection on the listener
func (srv *Server) Close() {
	srv.listener.Close()
}

func (srv *Server) handleAuth(conn net.Conn) error {
	var auth Packet
	err := auth.DecodeBinary(conn)
	if err != nil {
		return err
	}
	if auth.Header.Type != PacketTypeAuth {
		return fmt.Errorf("Unexpected packet type: %v", auth.Header.Type)
	}
	if auth.Body != srv.password {
		resp := Packet{
			Header: PacketHeader{
				ID:   PacketIDInvalid,
				Type: AuthResponsePacket,
			},
		}
		return resp.EncodeBinary(conn)
	}
	resp := Packet{
		Header: PacketHeader{
			ID:   auth.Header.ID,
			Type: AuthResponsePacket,
		},
	}
	return resp.EncodeBinary(conn)
}

// The HandlerFunc wraps functions to implement an RConHandler
type HandlerFunc func(cb ResponseCallback, cmd string) error

// ServeRCon calls f(cb, cmd)
func (f HandlerFunc) ServeRCon(cb ResponseCallback, cmd string) error {
	return f(cb, cmd)
}

// Handle registers a new connection handler, overwriting any existing one
func (srv *Server) Handle(cmd string, handler HandlerFunc) {
	if handler == nil {
		panic("nil handler")
	}
	srv.handlers[cmd] = handler
}

// HandleFunc registers a new connection handler from a function
func (srv *Server) HandleFunc(cmd string, handlerFunc HandlerFunc) {
	srv.Handle(cmd, HandlerFunc(handlerFunc))
}

func (srv *Server) handlePacket(conn net.Conn) error {
	var req Packet
	err := req.DecodeBinary(conn)
	if err != nil {
		return err
	}
	// Check for invalid message types
	if req.Header.Type != PacketTypeData {
		resp := Packet{
			Header: PacketHeader{
				ID:   req.Header.ID,
				Type: PacketTypeData,
			},
			Body: fmt.Sprintf("Unknown request %d", req.Header.Type),
		}
		return resp.EncodeBinary(conn)
	}
	// Find the handler
	fields := strings.Fields(req.Body)
	if len(fields) == 0 {
		return fmt.Errorf("no command found in body")
	}
	cmd := fields[0]
	handler, ok := srv.handlers[cmd]
	if !ok {
		return fmt.Errorf("no handler found for command: %s", cmd)
	}

	return handler.ServeRCon(func(response string) error {
		if len(response) > math.MaxInt32 {
			return fmt.Errorf("reponse too long")
		}
		length := int32(len(response))
		resp := Packet{
			Header: PacketHeader{
				ID:   req.Header.ID,
				Type: PacketTypeDataResponse,
			},
		}
		for {
			curLength := length
			if curLength > PacketMaxSize {
				curLength = PacketMaxSize
			}
			length -= curLength
			resp.Body = response[0:curLength]
			response = response[curLength:]
			err := resp.EncodeBinary(conn)
			if err != nil {
				return err
			}
			if length == 0 {
				return nil
			}
		}
	}, req.Body)
}

func (srv *Server) handleConnection(conn net.Conn) {
	err := srv.handleAuth(conn)
	if err != nil {
		log.Println("Error handling auth:", err)
		conn.Close()
		return
	}
	for {
		err := srv.handlePacket(conn)
		if err != nil {
			log.Println("Error handling packet:", err)
			conn.Close()
			return
		}
	}
}
