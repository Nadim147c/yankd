package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cast"
)

type pauseState int

const (
	PauseFalse pauseState = iota - 1
	PauseToggle
	PauseTrue
)

func SetPause(ctx context.Context, state pauseState) (bool, error) {
	c := NewClient()
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/pause/%d", BaseURL, state), nil)
	if err != nil {
		return false, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, extractError(resp)
	}

	var newState bool

	err = json.NewDecoder(resp.Body).Decode(&newState)

	return newState, err
}

func (s *Server) PauseHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state := pauseState(cast.ToInt(r.PathValue("state")))
		if state == PauseToggle {
			v := s.cb.TogglePaused()
			renderJSON(w, v)
			return
		}

		v := s.cb.SetPaused(state == PauseTrue)
		renderJSON(w, v)
	})
}
