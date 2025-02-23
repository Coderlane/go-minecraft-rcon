package rcon

import (
	"context"
	"fmt"
	"math"
	"net"
	"time"

	"golang.org/x/time/rate"
)

var (
	// MaxRequestsPerSecond represents the maximum number of requests per second
	// to send
	MaxRequestsPerSecond rate.Limit = 1
	// MaxParallelRequests represents the maximum number of parallel requests
	MaxParallelRequests = 1

	// DefaultTimeout represents the default timeout for messages
	DefaultTimeout time.Duration = time.Second * 25
)

// Conn represents a connection with an rcon server
type Conn struct {
	packetID int32
	conn     net.Conn

	limiter *rate.Limiter
}

// Dial dials the server and authenticates with the given password
func Dial(address, password string) (*Conn, error) {
	var err error
	conn := &Conn{
		limiter: rate.NewLimiter(MaxRequestsPerSecond, MaxParallelRequests),
	}
	conn.conn, err = net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	request := Packet{
		Header: PacketHeader{
			Type: PacketTypeAuth,
		},
		Body: password,
	}
	_, err = conn.request(request)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (conn *Conn) nextID() int32 {
	id := conn.packetID
	if conn.packetID != math.MaxInt32 {
		conn.packetID++
	} else {
		conn.packetID = 1
	}
	return id
}

func (conn *Conn) send(req Packet) (int32, error) {
	conn.conn.SetDeadline(time.Now().Add(DefaultTimeout))
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	err := conn.limiter.Wait(ctx)
	if err != nil {
		return -1, err
	}
	req.Header.ID = conn.nextID()
	return req.Header.ID, req.EncodeBinary(conn.conn)
}

func (conn *Conn) recv() (Packet, error) {
	var resp Packet
	err := resp.DecodeBinary(conn.conn)
	return resp, err
}

func (conn *Conn) request(pkt Packet) (string, error) {
	id, err := conn.send(pkt)
	if err != nil {
		return "", err
	}

	respBody := ""
	endID := PacketIDInvalid
	for {
		resp, err := conn.recv()
		if err != nil {
			return "", err
		}
		if resp.Header.ID == PacketIDInvalid {
			return "", fmt.Errorf("auth error")
		}
		if resp.Header.ID == endID {
			break
		}
		if resp.Header.ID != id {
			return "", fmt.Errorf("mismatched response. expected: %d != got: %d",
				resp.Header.ID, id)
		}
		respBody += resp.Body
		if len(resp.Body) < int(PacketMaxSize) &&
			endID == PacketIDInvalid {
			break
		} else if endID == PacketIDInvalid {
			// After the first packet with the maximum size, send a message
			// so we can figure out when the current message is done. We send an
			// invalid message that generates an error, but does not close the
			// connection.
			endPacket := Packet{
				Header: PacketHeader{
					Type: PacketTypeDataResponse,
				},
			}
			endID, err = conn.send(endPacket)
			if err != nil {
				return "", nil
			}
		}
	}
	return respBody, nil
}

// Request sends a request to the server and returns the response
func (conn *Conn) Request(body string) (string, error) {
	pkt := Packet{
		Header: PacketHeader{
			Type: PacketTypeData,
		},
		Body: body,
	}
	return conn.request(pkt)
}

// Send sends a request to the server and ignores the response
func (conn *Conn) Send(body string) error {
	pkt := Packet{
		Header: PacketHeader{
			Type: PacketTypeData,
		},
		Body: body,
	}
	_, err := conn.send(pkt)
	return err
}

// Close closes the connection to the server
func (conn *Conn) Close() error {
	return conn.conn.Close()
}
