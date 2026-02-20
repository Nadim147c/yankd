package ipc

import (
	"path/filepath"
	"testing"
	"time"
)

func TestIpcEcho(t *testing.T) {
	file := filepath.Join(t.TempDir(), SocketName)
	internalSocketPathOnlyForTesting = file
	server := NewServer(nil, nil)

	go server.Listen(t.Context())

	time.Sleep(50 * time.Millisecond)

	client := NewClient()
	if err := client.Connect(); err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer client.Close()

	msg := "hello world"

	resp, err := client.SendEcho(msg)
	if err != nil {
		t.Fatalf("SendEcho failed: %v", err)
	}

	if resp != msg {
		t.Fatalf("expected %q, got %q", msg, resp)
	}
}

func TestIpcPing(t *testing.T) {
	file := filepath.Join(t.TempDir(), SocketName)
	internalSocketPathOnlyForTesting = file
	server := NewServer(nil, nil)

	go server.Listen(t.Context())

	time.Sleep(50 * time.Millisecond)

	client := NewClient()
	if err := client.Connect(); err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer client.Close()

	oneWay, roundTrip, err := client.SendPing()
	if err != nil {
		t.Fatalf("SendEcho failed: %v", err)
	}

	t.Log("one-way latency", oneWay)
	t.Log("round-trip latency", roundTrip)
}
