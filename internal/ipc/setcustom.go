package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Nadim147c/yankd/internal/clipboard"
	"github.com/Nadim147c/yankd/internal/models"
)

// SetCustom sends a custom clipboard event to the daemon over IPC.
func SetCustom(ctx context.Context, event models.ClipboardEvent) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(event); err != nil {
		return err
	}

	c := NewClient()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, BaseURL+"/set-custom", &buf)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "Application/json")

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to ipc server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return extractError(resp)
	}
	return nil
}

// SetCustomHandler handles POST /set-custom requests from clients.
func (s *Server) SetCustomHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var event models.ClipboardEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(event.Entries) == 0 {
			http.Error(w, "no entries", http.StatusBadRequest)
			return
		}

		// Fill in missing fields and compute hash
		event.Time = time.Now()
		for _, e := range event.Entries {
			if strings.HasPrefix(e.MimeType, "image/") {
				event.MimeType = e.MimeType
				break
			}
		}
		if event.MimeType == "" && len(event.Entries) > 0 {
			event.MimeType = event.Entries[0].MimeType
		}
		for i := range event.Entries {
			event.Entries[i].Hash = models.NewHash(event.Entries[i].Blob)
			event.Entries[i].IsText = utf8.Valid(event.Entries[i].Blob)
		}
		event.Preview = clipboard.GeneratePreview(event.Entries)

		if err := s.cb.SetClipboard(event); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Save to history asynchronously to avoid blocking the Wayland event loop
		clone := event
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := s.db.Insert(ctx, &clone); err != nil {
				slog.Error("failed to save custom clipboard to history", "error", err)
			}
		}()

		renderJSON(w, map[string]string{"status": "ok"})
	})
}
