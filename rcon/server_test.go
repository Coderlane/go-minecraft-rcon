package rcon

import (
	"testing"
)

// Most test are already handled in connection_test.go

func TestListenInvalid(t *testing.T) {
	_, err := Listen("foo", "password")
	if err == nil {
		t.Fatal("Expected to fail")
	}
	expectError(t, err, "missing port")
}

func TestListenClose(t *testing.T) {
	srv, err := Listen("", "password")
	if err != nil {
		t.Fatal(err)
	}
	srv.Close()
}
