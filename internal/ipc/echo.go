package ipc

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

func SendEcho(ctx context.Context, msg string) (string, error) {
	c := NewClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, BaseURL+"/echo", strings.NewReader(msg))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "plain/text")

	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(resp.Body)
	return buf.String(), err
}

func (s *Server) EchoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		w.WriteHeader(http.StatusOK)
		_, err := io.Copy(w, r.Body)
		if err != nil {
			slog.Error("failed to copy request.Body to ResponseWriter", "error", err)
		}
	})
}
