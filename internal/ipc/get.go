package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Nadim147c/yankd/internal/models"
	"github.com/google/uuid"
)

func GetEvent(ctx context.Context, id uuid.UUID) (e models.ClipboardEvent, err error) {
	c := NewClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/get/%s", BaseURL, id.String()), nil)
	if err != nil {
		return e, err
	}
	req.Header.Add("Accept", "Application/json")

	resp, err := c.Do(req)
	if err != nil {
		return e, fmt.Errorf("failed to send request to ipc server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e, extractError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&e)
	return e, err
}

func (s *Server) GetEventHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := s.db.Get(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		renderJSON(w, res)
	})
}

func GetManyEvents(ctx context.Context, ids ...uuid.UUID) (e []models.ClipboardEvent, err error) {
	buf := bytes.NewBuffer(nil)
	_ = json.NewEncoder(buf).Encode(ids)
	c := NewClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, BaseURL+"/get", buf)
	if err != nil {
		return e, err
	}
	req.Header.Add("Content-Type", "Application/json")
	req.Header.Add("Accept", "Application/json")

	resp, err := c.Do(req)
	if err != nil {
		return e, fmt.Errorf("failed to send request to ipc server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e, extractError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(&e)
	return e, err
}

func (s *Server) GetManyEventsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var ids []uuid.UUID
		err := json.NewDecoder(r.Body).Decode(&ids)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := s.db.GetMany(r.Context(), ids...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		renderJSON(w, res)
	})
}
