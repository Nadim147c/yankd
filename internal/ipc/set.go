package ipc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

func (s *Server) handleSet(req *msg) *msg {
	var id int64
	_, err := fmt.Fscanf(req, "%d", &id)
	if err != nil {
		return newErrMsg(err)
	}

	event, err := s.db.Get(s.ctx, id)
	if err != nil {
		return newErrMsg(err)
	}

	slog.Debug("setting/offering clipboard!")
	err = s.cb.SetClipboard(event)
	if err != nil {
		return newErrMsg(err)
	}

	return new(msg)
}

func (c *Client) SendSet(ctx context.Context, id int64) error {
	req := new(msg)
	req.command = commandSet
	fmt.Fprint(req, id) //nolint:errcheck // writing small data with buffer
	resp, err := c.sendMsg(ctx, req)
	if err != nil {
		return err
	}
	if resp.status != statusOk {
		slog.Error("failed to set clipboard", "error", resp.String())
		return errors.New(resp.String())
	}
	return nil
}
