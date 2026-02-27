package ipc

import (
	"errors"
	"fmt"
	"log/slog"
)

type pauseCmd uint

const (
	PauseCmdFalse pauseCmd = iota
	PauseCmdTrue
	PauseCmdToggle
)

func (s *Server) handlePause(req *msg) *msg {
	var cmd pauseCmd
	_, err := fmt.Fscanf(req, "%d", &cmd)
	if err != nil {
		return newErrMsg(err)
	}

	if cmd < PauseCmdToggle {
		v := s.cb.SetPaused(cmd == PauseCmdTrue)
		m := new(msg)
		fmt.Fprint(m, v)
		return m
	}

	if cmd == PauseCmdToggle {
		v := s.cb.TogglePaused()
		m := new(msg)
		fmt.Fprint(m, v)
		return m
	}

	var m msg
	m.status = statusErr
	m.SetString("invalid command")

	return new(msg)
}

func (c *Client) SendPause(cmd pauseCmd) error {
	req := new(msg)
	req.command = commandPause
	fmt.Fprint(req, cmd)
	resp, err := c.sendMsg(req)
	if err != nil {
		return err
	}
	if resp.status != statusOk {
		return errors.New(resp.String())
	}
	slog.Info("history pause state has changed", "is_paused", resp.String())
	return nil
}
