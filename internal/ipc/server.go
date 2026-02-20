package ipc

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nadim147c/yankd/internal/db"
)

const (
	MaxBufferSize = 100 * 1024 * 1024
	SocketName    = "yankd.sock" // default socket file name
)

var internalSocketPathOnlyForTesting string

// getSocketPath returns the full path to the socket using XDG_RUNTIME_DIR or /tmp.
func getSocketPath() string {
	if internalSocketPathOnlyForTesting != "" {
		return internalSocketPathOnlyForTesting
	}

	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		return filepath.Join(runtimeDir, SocketName)
	}
	return filepath.Join("/tmp", SocketName)
}

// Server listens on a Unix socket, accepts connections, and processes messages.
type Server struct {
	listener   *net.UnixListener
	socketPath string
}

var ErrAlreadyRunning = errors.New("another instance is running")

// CreateServer creates and starts listening on the automatically determined socket path.
// It removes any existing socket file to avoid address already in use errors.
func CreateServer() (*Server, error) {
	socketPath := getSocketPath()

	if _, err := os.Stat(socketPath); err == nil {
		return nil, ErrAlreadyRunning
	}

	addr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve unix addr: %w", err)
	}

	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on unix socket: %w", err)
	}

	return &Server{listener: listener, socketPath: socketPath}, nil
}

// Listen accepts incoming connections and handles them until the context is
// cancelled. The db parameter is currently ignored (reserved for future use).
func (s *Server) Listen(ctx context.Context, db *db.DB) error {
	defer s.listener.Close()

	// Channel to signal when accept loop should stop
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		_ = s.listener.Close() // this will break the Accept loop
		close(done)
	}()

	for {
		conn, err := s.listener.AcceptUnix()
		if err != nil {
			// If the listener was closed due to context cancellation, return nil
			// (normal shutdown)
			select {
			case <-ctx.Done():
				return nil
			default:
				return fmt.Errorf("accept error: %w", err)
			}
		}

		// Handle each connection concurrently
		go s.handleConnection(conn, db)
	}
}

// handleConnection reads messages from a single connection and responds.
func (s *Server) handleConnection(conn *net.UnixConn, db *db.DB) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(nil, MaxBufferSize)

	for scanner.Scan() {
		line := scanner.Text() // reads up to newline
		cmd, payload, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		cmd = strings.TrimSpace(cmd)

		switch cmd {
		case "echo":
			s.HandleEcho(conn, payload, db)
		case "ping":
			s.HandlePing(conn, payload, db)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "connection read error: %v\n", err)
	}
}

// Close closes the server's listener. It does NOT close any database (per comment).
func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		return err
	}
	return os.Remove(s.socketPath)
}
