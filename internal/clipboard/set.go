package clipboard

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"

	"github.com/Nadim147c/yankd/internal/models"
	protocol "github.com/Nadim147c/yankd/internal/wlr-data-control-unstable-v1"
)

// YankdMimeType mime type is used for share yankd clipboard metadata.
const YankdMimeType = "application/yankd"

func (c *Client) HandleZwlrDataControlSourceV1Send(e protocol.ZwlrDataControlSourceV1SendEvent) {
	f := os.NewFile(e.Fd, "")
	if f == nil {
		return
	}
	defer f.Close()

	if c.event == nil {
		return
	}

	if e.MimeType == YankdMimeType {
		if err := json.NewEncoder(f).Encode(c.event); err != nil {
			slog.Error("failed to encode yankd metadata", "error", err)
		}
		return
	}

	data, ok := c.mimes[e.MimeType]
	if !ok {
		return
	}

	_, err := f.Write(data)
	if err != nil {
		slog.Error("failed to encode yankd metadata", "error", err)
	}
}

func (c *Client) SetClipboard(event models.ClipboardEvent) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	slog.Debug("set clipboard has been requested")

	if !c.connected.Load() {
		return errors.New("client is not connected to wayland server")
	}

	src, err := c.manager.CreateDataSource()
	if err != nil {
		return err
	}

	m := make(map[string][]byte, len(event.Entries))
	src.AddSendHandler(c)
	for _, entry := range event.Entries {
		if err := src.Offer(entry.MimeType); err != nil {
			return err
		}
		if entry.IsText {
			m[entry.MimeType] = entry.Text.Bytes()
		} else {
			m[entry.MimeType] = entry.Blob
		}
	}

	// This is so that we can know if the clipboard is set by yankd
	if err := src.Offer(YankdMimeType); err != nil {
		return err
	}

	if err := c.device.SetSelection(src); err != nil {
		return err
	}

	c.mimes = m
	event.Entries = nil
	c.event = &event

	return nil
}
