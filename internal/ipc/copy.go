package ipc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"

	"github.com/Nadim147c/yankd/internal/models"
)

func encode(v any) string {
	p, err := json.Marshal(v)
	if err != nil {
		// In no circumstance, we need to send data that can't be encoded!
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(p)
}

func decode(p string, v any) error {
	data, err := base64.StdEncoding.DecodeString(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (s *Server) HandleCopy(conn *net.UnixConn, payload string) {
	var event models.ClipboardEvent
	if err := decode(payload, &event); err != nil {
		fmt.Fprintln(conn, BadRequest) //nolint:errcheck
	}
	fmt.Fprintln(conn, Success) //nolint:errcheck
}

func (c *Client) SendCopy(event models.ClipboardEvent) (string, error) {
	resp, err := c.Send("copy", encode(event))
	return string(resp), err
}
