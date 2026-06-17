package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func DeteteEvents(ctx context.Context, ids ...uuid.UUID) (n int64, err error) {
	buf := bytes.NewBuffer(nil)
	json.NewEncoder(buf).Encode(ids)
	c := NewClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, BaseURL+"/delete", buf)
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

func (s *Server) DeteteEventsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var ids []uuid.UUID
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := s.db.Delete(r.Context(), ids...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		renderJSON(w, res)
	})
}
