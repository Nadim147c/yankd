package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/Nadim147c/yankd/internal/clipboard"
	"github.com/Nadim147c/yankd/internal/models"
)

func SetClipboard(ctx context.Context, payload map[string][]byte) (models.ClipboardEvent, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return models.ClipboardEvent{}, err
	}

	c := NewClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, BaseURL+"/copy", bytes.NewReader(buf))
	if err != nil {
		return models.ClipboardEvent{}, err
	}

	req.Header.Add("Accept", "Application/json")
	req.Header.Add("Content-Type", "Application/json")

	resp, err := c.Do(req)
	if err != nil {
		return models.ClipboardEvent{}, fmt.Errorf("failed to send request to ipc server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.ClipboardEvent{}, extractError(resp)
	}

	var e models.ClipboardEvent
	err = json.NewDecoder(resp.Body).Decode(&e)
	if err != nil {
		return models.ClipboardEvent{}, err
	}
	return e, nil
}

func (s *Server) SetClipboardHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var payload map[string][]byte
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var event models.ClipboardEvent

		event.Time = time.Now()

		mimes := make([]string, 0, len(payload))
		entries := make([]models.ClipboardEntry, 0, len(payload))
		for mime, v := range payload {
			var entry models.ClipboardEntry
			entry.Hash = models.NewHash(v)
			entry.MimeType = mime
			entry.IsText = utf8.Valid(v)
			entry.Blob = v
			mimes = append(mimes, mime)
			entries = append(entries, entry)
		}

		event.Entries = entries
		event.Preview = clipboard.GeneratePreivew(entries)
		event.MimeType = clipboard.SelectMimeType(mimes)

		err = s.db.Insert(r.Context(), &event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = s.cb.SetClipboard(event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		renderJSON(w, event)
	})
}
