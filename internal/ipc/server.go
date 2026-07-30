package ipc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Nadim147c/yankd/internal/clipboard"
	"github.com/Nadim147c/yankd/internal/db"
)

const (
	BaseURL    = "http://unix"
	SocketName = "yankd.sock" // default socket file name
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
	db  *db.DB
	cb  *clipboard.Client
	ctx context.Context
}

var ErrAlreadyRunning = errors.New("another instance is running")

// NewServer creates and starts listening on the automatically determined socket path.
// It removes any existing socket file to avoid address already in use errors.
func NewServer(db *db.DB, cb *clipboard.Client) *Server {
	s := new(Server)
	s.db = db
	s.cb = cb
	return s
}

func loggingMiddleware(next http.Handler) http.Handler {
	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Debug("new request", "method", r.Method, "path", r.URL.Path)

			start := time.Now()
			next.ServeHTTP(w, r)

			slog.Debug("request served", "method", r.Method, "path", r.URL.Path, "took", time.Since(start))
		})
	}
	return next
}

// Listen accepts incoming connections and handles them until the context is
// cancelled. The db parameter is currently ignored (reserved for future use).
func (s *Server) Listen(ctx context.Context) error {
	socketPath := getSocketPath()
	s.ctx = ctx
	defer func() { s.ctx = nil }()

	if _, err := os.Stat(socketPath); err == nil {
		return ErrAlreadyRunning
	}
	defer os.Remove(socketPath)

	addr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to resolve unix addr: %w", err)
	}

	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %w", err)
	}
	defer listener.Close()
	slog.Info("Listening for ipc request", "addr", addr)

	mux := http.NewServeMux()
	mux.Handle("GET /get/{id}", s.GetEventHandler())
	mux.Handle("GET /search", s.SearchHandler())
	mux.Handle("POST /delete", s.DeteteEventsHandler())
	mux.Handle("POST /echo", s.EchoHandler())
	mux.Handle("POST /get", s.GetManyEventsHandler())
	mux.Handle("POST /pause/{state}", s.PauseHandler())
	mux.Handle("POST /set/{id}", s.SetEventHandler())
	mux.Handle("POST /wipe", s.WipeDatabaseHandler())

	handler := loggingMiddleware(mux)

	httpServer := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serverError := make(chan error, 1)

	go func() {
		if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError <- fmt.Errorf("http server error: %w", err)
		}
		close(serverError)
	}()

	select {
	case <-ctx.Done():
		// Context was cancelled, gracefully shut down the server
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("failed to gracefully shutdown: %w", err)
		}
		return ctx.Err() // returns context.Canceled
	case err := <-serverError:
		return err
	}
}
