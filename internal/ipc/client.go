package ipc

import "net"

// Client connects to a Unix socket and sends/receives messages.
type Client struct{}

func NewClient() *Client {
	return new(Client)
}

func (c *Client) send(cmd command, data []byte) (*msg, error) {
	var req msg
	req.command = cmd
	req.buf = data
	return c.sendMsg(&req)
}

func (c *Client) sendMsg(req *msg) (*msg, error) {
	socketPath := getSocketPath()

	conn, err := net.Dial("unix", socketPath)
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
