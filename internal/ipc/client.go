package ipc

import (
	"context"
	"net"
	"net/http"
)

// Client connects to a Unix socket and sends/receives messages.
type Client struct{}

func NewClient() *http.Client {
	socket := getSocketPath()
	return &http.Client{ //nolint:exhaustruct
		Transport: &http.Transport{ //nolint:exhaustruct
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "unix", socket)
			},
		},
	}
}
