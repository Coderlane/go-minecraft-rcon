package client

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	testPassword string = "hunter2"
)

var (
	testServerAddress string = ""
)

type testServer struct {
	listener net.Listener
}

func newTestServer() (*testServer, error) {
	listener, err := net.Listen("tcp", testServerAddress)
	if err != nil {
		return nil, err
	}
	return &testServer{
		listener: listener,
	}, nil
}

func (srv *testServer) Serve() string {
	go func() {
		for {
			conn, err := srv.listener.Accept()
			if err != nil {
				return
			}
			conn.SetDeadline(time.Time{})
			go handleConnection(conn)
		}
	}()

	return srv.listener.Addr().String()
}

func (srv *testServer) Close() {
	srv.listener.Close()
}

func handleAuth(conn net.Conn) error {
	var auth Packet
	err := auth.DecodeBinary(conn)
	if err != nil {
		return err
	}
	if auth.Header.Type != AuthPacket {
		return fmt.Errorf("Unexpected packet type: %v", auth.Header.Type)
	}
	if auth.Body != testPassword {
		resp := Packet{
			Header: PacketHeader{
				ID:   invalidMessageID,
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

func handleData(conn net.Conn) error {
	var req Packet
	err := req.DecodeBinary(conn)
	if err != nil {
		return err
	}
	if req.Header.Type != DataPacket {
		return fmt.Errorf("Unexpected packet type: %v", req.Header.Type)
	}
	// Don't ack snd messages
	if req.Body == "snd" {
		return nil
	}
	var length int32
	_, err = fmt.Sscanf(req.Body, "msg-%d", &length)
	if err != nil {
		return err
	}
	resp := Packet{
		Header: PacketHeader{
			ID:   req.Header.ID,
			Type: DataPacketResponse,
		},
	}
	for {
		curLength := length
		if curLength > packetMaxSize {
			curLength = packetMaxSize
		}
		length -= curLength
		body := make([]byte, curLength)
		for i := int32(0); i < curLength; i++ {
			body[i] = 'a'
		}
		resp.Body = string(body)
		err := resp.EncodeBinary(conn)
		if err != nil {
			return err
		}
		if length == 0 {
			break
		}
	}
	return nil
}

func handleConnection(conn net.Conn) {
	err := handleAuth(conn)
	if err != nil {
		fmt.Println("Error handling auth:", err)
		conn.Close()
		return
	}
	for {
		err := handleData(conn)
		if err != nil {
			fmt.Println("Error handling packet:", err)
			conn.Close()
			return
		}
	}
}

func expectError(t *testing.T, err error, contains string) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), contains) {
		t.Errorf("Expected \"%s\": %v", contains, err)
	}
}

func TestAuthSuccess(t *testing.T) {
	_, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAuthFailure(t *testing.T) {
	_, err := Dial(testServerAddress, testPassword+"incorrect")
	if err == nil {
		t.Fatal("Expected to fail")
	}
	expectError(t, err, "auth error")
}

func TestAddrInvalid(t *testing.T) {
	_, err := Dial("-1", testPassword)
	if err == nil {
		t.Fatal("Expected to fail")
	}
	expectError(t, err, "dial tcp")
}

func TestRequestData(t *testing.T) {
	tcases := []int32{0, 1, packetMaxSize - 1, packetMaxSize, packetMaxSize + 1,
		3*packetMaxSize + 1}
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}

	for _, tcase := range tcases {
		t.Run(fmt.Sprintf("Size-%d", tcase), func(t *testing.T) {
			resp, err := rconn.Request(fmt.Sprintf("msg-%d", tcase))
			if err != nil {
				t.Error(err)
			}
			if int32(len(resp)) != tcase {
				t.Errorf("Expected length: %d got: %d", tcase, len(resp))
			}
		})
	}
}

func TestRequestInvalid(t *testing.T) {
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rconn.Request("invalid")
	expectError(t, err, "EOF")
}

func TestRequestOutOfOrder(t *testing.T) {
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	err = rconn.Send("msg-1")
	if err != nil {
		t.Error(err)
	}
	_, err = rconn.Request("msg-1")
	expectError(t, err, "mismatched response")
}

func TestRequestOnClosedConnection(t *testing.T) {
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	rconn.Close()
	_, err = rconn.Request("msg-1")
	expectError(t, err, "closed network connection")
}

func TestSend(t *testing.T) {
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	err = rconn.Send("snd")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendOnClosedConnection(t *testing.T) {
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	rconn.Close()
	err = rconn.Send("snd")
	expectError(t, err, "closed network connection")
}

func TestSendWouldTimeout(t *testing.T) {
	save := MaxRequestsPerSecond
	MaxRequestsPerSecond = 0.000001
	defer func() {
		MaxRequestsPerSecond = save
	}()

	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	err = rconn.Send("snd")
	expectError(t, err, "would exceed context deadline")
}

func TestMain(m *testing.M) {
	FragmentMessage = "msg-1"
	MaxRequestsPerSecond = 100

	srv, err := newTestServer()
	if err != nil {
		fmt.Println("Failed to listen:", err)
		os.Exit(1)
	}
	testServerAddress = srv.Serve()
	code := m.Run()
	srv.Close()
	os.Exit(code)
}
