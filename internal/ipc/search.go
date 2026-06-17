package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/spf13/cast"
)

func GetSearch(ctx context.Context, query string, limit int64) ([]db.SearchResult, error) {
	c := NewClient()
	req, err := http.NewRequest(http.MethodGet, BaseURL+"/search", nil)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("query", query)
	q.Set("limit", cast.ToString(limit))
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Accept", "Application/json")

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to ipc server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, extractError(resp)
	}

	var res []db.SearchResult
	err = json.NewDecoder(resp.Body).Decode(&res)
	return res, err
}

func (s *Server) SearchHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		q := r.URL.Query()
		query := q.Get("query")
		limit := cast.ToInt64(q.Get("limit"))

		if query == "" {
			items, err := s.db.List(r.Context(), limit)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			renderJSON(w, items)
			return
		}

		items, err := s.db.Search(r.Context(), query, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		renderJSON(w, items)
	})
}
