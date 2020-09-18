package client

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Coderlane/go-minecraft-rcon/rcon"
)

const (
	testPassword string = "password"
)

var (
	testClient        Client
	testServerAddress string
)

func expectError(t *testing.T, err error, expected string) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected: %s Got: %v", expected, err)
	}
}

func TestInvalidAddress(t *testing.T) {
	_, err := NewClient("invalid", testPassword)
	if err == nil {
		t.Fatal("Expected an error")
	}
}

func TestRequest(t *testing.T) {
	resp, err := testClient.Request("req")
	if err != nil {
		t.Fatal(err)
	}
	if resp != "resp" {
		t.Error("Expected \"resp\":", resp)
	}
}

func TestRequestInvalid(t *testing.T) {
	_, err := testClient.Request("req\nreq")
	expectError(t, err, "invalid command: req")
}

func TestSend(t *testing.T) {
	err := testClient.Send("snd")
	if err != nil {
		t.Fatal(err)
	}
}

func TestSendInvalid(t *testing.T) {
	err := testClient.Send("snd\nsnd")
	expectError(t, err, "invalid command: snd")
}

func TestClose(t *testing.T) {
	client, err := NewClient(testServerAddress, testPassword)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Send("snd")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestMain(m *testing.M) {
	rcon.MaxRequestsPerSecond = 100
	srv, err := rcon.Listen("", testPassword)
	if err != nil {
		fmt.Println("Failed to create server:", err)
		os.Exit(1)
	}
	srv.HandleFunc("snd", func(cb rcon.ResponseCallback, cmd string) error {
		return nil
	})
	srv.HandleFunc("req", func(cb rcon.ResponseCallback, cmd string) error {
		return cb("resp")
	})
	testServerAddress = srv.Addr().String()
	testClient, err = NewClient(testServerAddress, testPassword)
	code := m.Run()
	srv.Close()
	os.Exit(code)
}
