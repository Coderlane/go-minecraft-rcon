package rcon

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
)

const (
	testPassword string = "hunter2"
)

var (
	testServerAddress string = ""
)

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

func TestDialInvalid(t *testing.T) {
	_, err := Dial("-1", testPassword)
	if err == nil {
		t.Fatal("Expected to fail")
	}
	expectError(t, err, "dial tcp")
}

func TestRequestData(t *testing.T) {
	tcases := []int32{0, 1, PacketMaxSize - 1, PacketMaxSize, PacketMaxSize + 1,
		3*PacketMaxSize + 1}
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}

	for _, tcase := range tcases {
		t.Run(fmt.Sprintf("Size-%d", tcase), func(t *testing.T) {
			resp, err := rconn.Request(fmt.Sprintf("req %d", tcase))
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

func TestRequestEmpty(t *testing.T) {
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rconn.Request("")
	expectError(t, err, "EOF")
}

func TestRequestOutOfOrder(t *testing.T) {
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	err = rconn.Send("req 1")
	if err != nil {
		t.Error(err)
	}
	_, err = rconn.Request("req 1")
	expectError(t, err, "mismatched response")
}

func TestRequestOnClosedConnection(t *testing.T) {
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	rconn.Close()
	_, err = rconn.Request("req 1")
	expectError(t, err, "closed network connection")
}

func TestRequestGenTooLarge(t *testing.T) {
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	_, err = rconn.Request(fmt.Sprintf("req %d", math.MaxInt32+1))
	expectError(t, err, "EOF")
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

func TestSendThenRequest(t *testing.T) {
	rconn, err := Dial(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	err = rconn.Send("snd")
	if err != nil {
		t.Fatal(err)
	}
	_, err = rconn.Request("req 1")
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
	MaxRequestsPerSecond = 100
	srv, err := Listen("", testPassword)
	if err != nil {
		fmt.Println("Failed to create server:", err)
		os.Exit(1)
	}

	srv.HandleFunc("snd", func(cb ResponseCallback, cmd string) error {
		return nil
	})
	srv.HandleFunc("req", func(cb ResponseCallback, cmd string) error {
		var length int
		fmt.Sscanf(cmd, "req %d", &length)
		body := make([]byte, length)
		for i := 0; i < length; i++ {
			body[i] = 'a'
		}
		return cb(string(body))
	})

	testServerAddress = srv.Addr().String()
	code := m.Run()
	srv.Close()
	os.Exit(code)
}
