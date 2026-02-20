package ipc

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
)

// Client connects to a Unix socket and sends/receives messages.
type Client struct {
	conn   *net.UnixConn
	reader *bufio.Reader
	mu     sync.Mutex
}

func NewClient() *Client {
	return new(Client)
}

// Connect dials the automatically determined Unix socket path.
func (c *Client) Connect() error {
	socketPath := getSocketPath()
	addr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to resolve unix addr: %w", err)
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to dial unix socket: %w", err)
	}

	c.reader = bufio.NewReader(conn)
	c.conn = conn
	return nil
}

// Send writes a message in the format "<cmd>:<payload>" to the socket.
// It appends a newline automatically.
func (c *Client) Send(cmd string, args ...string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := fmt.Fprintf(c.conn, "%s:%s\n", cmd, strings.Join(args, ":"))
	if err != nil {
		return nil, err
	}

	line, prefix, err := c.reader.ReadLine()
	if err != nil {
		return nil, err
	}
	if !prefix {
		return line, nil
	}

	buf := bytes.NewBuffer(line)
	for prefix {
		line, prefix, err = c.reader.ReadLine()
		if err != nil {
			return nil, err
		}
		if buf.Len()+len(line) > MaxBufferSize {
			return nil, errors.New("max buffer size exceeded")
		}
		buf.Write(line)
	}

	return buf.Bytes(), nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
