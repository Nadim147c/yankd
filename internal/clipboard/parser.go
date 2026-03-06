package clipboard

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	mimepkg "mime"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Nadim147c/yankd/internal/models"
	protocol "github.com/Nadim147c/yankd/internal/wlr-data-control-unstable-v1"
	"github.com/gabriel-vasile/mimetype"
)

type clipboardParser struct {
	offer *protocol.ZwlrDataControlOfferV1
	mimes []string
}

// newClipboardParser creates a new parser for an offer.
func newClipboardParser(
	offer *protocol.ZwlrDataControlOfferV1,
	mimes []string,
) *clipboardParser {
	slog.Debug("creating clipboard parser", "offered_mimes_count", len(mimes))
	return &clipboardParser{offer, mimes}
}

// retrieveData fetches data for a specific MIME type.
func (c *clipboardParser) retrieveData(mimeType string) ([]byte, error) {
	slog.Debug("retrieving data", "mime", mimeType)

	if c.offer == nil {
		slog.Error("offer is nil", "mime", mimeType)
		return nil, errors.New("offer is nil")
	}

	// Create a pipe to receive data
	reader, writer, err := os.Pipe()
	if err != nil {
		slog.Error("failed to create pipe", "mime", mimeType, "error", err)
		return nil, fmt.Errorf("failed to create pipe for %s: %w", mimeType, err)
	}
	defer writer.Close()
	defer reader.Close()

	// Send receive request
	if err := c.offer.Receive(mimeType, writer.Fd()); err != nil {
		reader.Close() //nolint
		slog.Error("receive request failed", "mime", mimeType, "error", err)
		return nil, fmt.Errorf("receive request failed for %s: %w", mimeType, err)
	}

	// Close write end in this process
	writer.Close() //nolint

	// Read data from the read end
	data, err := io.ReadAll(reader)
	reader.Close() //nolint

	if err != nil {
		slog.Error("failed to read data", "mime", mimeType, "error", err)
		return nil, fmt.Errorf("failed to read data for %s: %w", mimeType, err)
	}

	return data, nil
}

// selectMime selects best mime for current clipboard item. Prefer image/* mimes
// with valid file extensions. Fallback: text/plain + ".txt".
func selectMime(m []string) (mime string) {
	// First pass: look for image/*
	for mt := range slices.Values(m) {
		mtype, _, _ := mimepkg.ParseMediaType(mt)
		if strings.HasPrefix(mtype, "image/") {
			return mt
		}
	}

	// Fallback: text/plain
	return "text/plain"
}

var imageRegex = regexp.MustCompile(`^(image/.+|.+/ico)$`)

// parse converts the retrieved data into a Clip struct.
func (c *clipboardParser) parse() (models.ClipboardEvent, error) {
	slog.Debug("parsing clipboard data")

	var event models.ClipboardEvent
	event.Time = time.Now()

	// Set MIME type
	event.PrimaryMimeType = selectMime(c.mimes)

	entries := make([]models.ClipboardEntry, 0, len(c.mimes))
	for mime := range slices.Values(c.mimes) {
		v, err := c.retrieveData(mime)
		if err != nil {
			return event, err
		}
		if len(v) == 0 {
			continue
		}

		hash := models.NewHash(v)
		mt, _, _ := mimepkg.ParseMediaType(mime)

		var entry models.ClipboardEntry
		entry.Hash = hash
		entry.MimeType = mime
		if imageRegex.MatchString(mt) {
			entry.Blob = v
		} else {
			entry.IsText = true
			entry.Text = models.NewNullString(string(v), true)
		}
		entries = append(entries, entry)
	}
	event.Entries = entries

	return event, nil
}

func getExt(b []byte) string {
	var ext string
	if mt := mimetype.Detect(b); mt != nil {
		ext = mt.Extension()
	}
	return ext
}
