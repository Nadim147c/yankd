package ipc

import (
	"fmt"
	"time"
)

func (s *Server) handleEcho(req *msg) *msg {
	return req
}

func (c *Client) SendEcho(msg string) (string, error) {
	resp, err := c.send(commandEcho, []byte(msg))
	return resp.String(), err
}

func (s *Server) handlePing(*msg) *msg {
	msg := new(msg)
	fmt.Fprint(msg, time.Now().UnixNano()) //nolint:errcheck
	return msg
}

func (c *Client) SendPing() (
	oneWayLatency time.Duration,
	roundTripLatency time.Duration,
	err error,
) {
	startTime := time.Now()

	resp, err := c.send(commandPing, nil)
	if err != nil {
		return 0, 0, err
	}

	var serverNano int64
	_, err = fmt.Fscanf(resp, "%d", &serverNano)
	if err != nil {
		return 0, 0, err
	}

	oneWayLatency = time.Duration(serverNano - startTime.UnixNano())
	roundTripLatency = time.Since(startTime)

	return oneWayLatency, roundTripLatency, nil
}
