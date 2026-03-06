package ipc

import (
	"path/filepath"
	"testing"
	"time"
)

func TestIpcEcho(t *testing.T) {
	internalSocketPathOnlyForTesting = filepath.Join(t.TempDir(), SocketName)
	server := NewServer(nil, nil)
	go server.Listen(t.Context())

	client := NewClient()
	msg := "hello world"

	time.Sleep(100 * time.Millisecond)

	resp, err := client.SendEcho(msg)
	if err != nil {
		t.Fatalf("SendEcho failed: %v", err)
	}

	if resp != msg {
		t.Fatalf("expected %q, got %q", msg, resp)
	}
}

func TestIpcPing(t *testing.T) {
	internalSocketPathOnlyForTesting = filepath.Join(t.TempDir(), SocketName)
	server := NewServer(nil, nil)
	go server.Listen(t.Context())

	client := NewClient()

	time.Sleep(100 * time.Millisecond)

	oneWay, roundTrip, err := client.SendPing()
	if err != nil {
		t.Fatalf("SendEcho failed: %v", err)
	}

	t.Log("one-way latency (server time)", oneWay)
	t.Log("round-trip latency", roundTrip)
}
