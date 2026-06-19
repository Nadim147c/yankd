package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func WipeDatabase(ctx context.Context) (n int64, err error) {
	c := NewClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, BaseURL+"/wipe", nil)
	if err != nil {
		return n, err
	}
	req.Header.Add("Content-Type", "Application/json")
	req.Header.Add("Accept", "Application/json")

	resp, err := c.Do(req)
	if err != nil {
		return n, fmt.Errorf("failed to send request to ipc server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return n, extractError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&n)
	return n, err
}

func (s *Server) WipeDatabaseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		res, err := s.db.Wipe(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		renderJSON(w, res)
	})
}
