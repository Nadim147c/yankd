package ipc

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/Nadim147c/yankd/internal/db"
)

const BadRequest = "<BAD_REQUEST>"

func (s *Server) HandleEcho(conn *net.UnixConn, payload string, db *db.DB) {
	_, _ = fmt.Fprintln(conn, payload)
}

func (c *Client) SendEcho(msg string) (string, error) {
	resp, err := c.Send("echo", msg)
	return string(resp), err
}

func (s *Server) HandlePing(conn *net.UnixConn, _ string, db *db.DB) {
	_, _ = fmt.Fprintln(conn, time.Now().UnixNano())
}

// SendPing sends a ping to the server and returns the estimated one-way
// latency (client → server) and the full round-trip latency.
func (c *Client) SendPing() (
	oneWayLatency time.Duration,
	roundTripLatency time.Duration,
	err error,
) {
	startTime := time.Now()

	resp, err := c.Send("ping", "")
	if err != nil {
		return 0, 0, err
	}

	serverNano, err := strconv.ParseInt(string(resp), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	oneWayLatency = time.Duration(serverNano - startTime.UnixNano())
	roundTripLatency = time.Since(startTime)

	return oneWayLatency, roundTripLatency, nil
}
