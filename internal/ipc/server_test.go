package ipc

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestEcho(t *testing.T) {
	tmpDir := t.TempDir()
	internalSocketPathOnlyForTesting = filepath.Join(tmpDir, "test.sock")

	s := NewServer(nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Listen(ctx)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	msg := "hello world"
	resp, err := SendEcho(ctx, msg)
	if err != nil {
		t.Fatalf("SendEcho failed: %v", err)
	}
	if resp != msg {
		t.Fatalf("expected %q, got %q", msg, resp)
	}

	cancel()
	err = <-errCh
	if err != nil && err != context.Canceled {
		t.Fatalf("Listen returned unexpected error: %v", err)
	}
}
