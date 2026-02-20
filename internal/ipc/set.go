package ipc

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"
)

func (s *Server) HandleSet(conn *net.UnixConn, payload string) {
	id, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		fmt.Fprintf(conn, "%s: %v\n", BadRequest, err) //nolint:errcheck
		return
	}
	event, err := s.db.Get(s.ctx, id)
	if err != nil {
		fmt.Fprintf(conn, "%s: %v\n", BadRequest, err) //nolint:errcheck
		return
	}
	slog.Debug("setting/offering clipboard!")
	err = s.cb.SetClipboard(event)
	if err != nil {
		fmt.Fprintf(conn, "%s: %v\n", BadRequest, err) //nolint:errcheck
		return
	}

	fmt.Fprintln(conn, Success) //nolint:errcheck
}

func (c *Client) SendSet(id int64) (string, error) {
	resp, err := c.Send("set", strconv.FormatInt(id, 10))
	return string(resp), err
}
