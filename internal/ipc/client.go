package ipc

import (
	"context"
	"net"
)

// Client connects to a Unix socket and sends/receives messages.
type Client struct{}

func NewClient() *Client {
	return new(Client)
}

func (c *Client) send(ctx context.Context, cmd command, data []byte) (*msg, error) {
	var req msg
	req.command = cmd
	req.buf = data
	return c.sendMsg(ctx, &req)
}

func (c *Client) sendMsg(ctx context.Context, req *msg) (*msg, error) {
	socketPath := getSocketPath()

	var dialer net.Dialer

	conn, err := dialer.DialContext(ctx, "unix", socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = req.Encode(conn)
	if err != nil {
		return nil, err
	}

	var resp msg
	return &resp, resp.Decode(conn)
}
