package client

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

	// FragmentMessage represents the message sent to determine the end of a
	// fragmented response from the server. Pick a message that has small
	// response.
	FragmentMessage string = "seed"
)

const (
	invalidMessageID int32 = -1
)

// RConn represents a connection with an rconn server
type RConn struct {
	packetID int32
	conn     net.Conn

	limiter *rate.Limiter
}

// Dial dials the server and authenticates with the given password
func Dial(address, password string) (RConn, error) {
	var err error
	rconn := RConn{
		limiter: rate.NewLimiter(MaxRequestsPerSecond, MaxParallelRequests),
	}
	rconn.conn, err = net.Dial("tcp", address)
	if err != nil {
		return rconn, err
	}
	request := Packet{
		Header: PacketHeader{
			Type: AuthPacket,
		},
		Body: password,
	}
	_, err = rconn.request(request)
	if err != nil {
		return rconn, err
	}
	return rconn, nil
}

func (rconn *RConn) nextID() int32 {
	id := rconn.packetID
	if rconn.packetID != math.MaxInt32 {
		rconn.packetID++
	} else {
		rconn.packetID = 1
	}
	return id
}

func (rconn *RConn) send(req Packet) (int32, error) {
	rconn.conn.SetDeadline(time.Now().Add(DefaultTimeout))
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	err := rconn.limiter.Wait(ctx)
	if err != nil {
		return -1, err
	}
	req.Header.ID = rconn.nextID()
	return req.Header.ID, req.EncodeBinary(rconn.conn)
}

func (rconn *RConn) recv() (Packet, error) {
	var resp Packet
	err := resp.DecodeBinary(rconn.conn)
	return resp, err
}

func (rconn *RConn) request(pkt Packet) (string, error) {
	id, err := rconn.send(pkt)
	if err != nil {
		return "", err
	}

	respBody := ""
	endID := invalidMessageID
	for {
		resp, err := rconn.recv()
		if err != nil {
			return "", err
		}
		if resp.Header.ID == invalidMessageID {
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
		if len(resp.Body) < int(packetMaxSize) &&
			endID == invalidMessageID {
			break
		} else if endID == invalidMessageID {
			// After the first packet with the maximum size, send a message
			// so we can figure out when the current message is done.
			endPacket := Packet{
				Header: PacketHeader{
					Type: DataPacket,
				},
				Body: FragmentMessage,
			}
			endID, err = rconn.send(endPacket)
			if err != nil {
				return "", nil
			}
		}
	}
	return respBody, nil
}

// Request sends a request to the server and returns the response
func (rconn *RConn) Request(body string) (string, error) {
	pkt := Packet{
		Header: PacketHeader{
			Type: DataPacket,
		},
		Body: body,
	}
	return rconn.request(pkt)
}

// Send sends a request to the server and ignores the response
func (rconn *RConn) Send(body string) error {
	pkt := Packet{
		Header: PacketHeader{
			Type: DataPacket,
		},
		Body: body,
	}
	_, err := rconn.send(pkt)
	return err
}

// Close closes the connection to the server
func (rconn *RConn) Close() error {
	return rconn.conn.Close()
}
