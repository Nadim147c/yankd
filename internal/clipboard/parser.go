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
	"unicode/utf8"

	"github.com/Nadim147c/yankd/internal/models"
	protocol "github.com/Nadim147c/yankd/internal/wlr-data-control-unstable-v1"
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
	for _, mt := range m {
		mtype, _, _ := mimepkg.ParseMediaType(mt)
		if strings.HasPrefix(mtype, "image/") {
			return mt
		}
	}

	// Fallback: text/plain
	return "text/plain"
}

var badMimes = []*regexp.Regexp{
	regexp.MustCompile(`text\/_moz_html`),
	regexp.MustCompile(`ico`),
	regexp.MustCompile(`BMP|bmp`),
	regexp.MustCompile(`bitmap`),
	regexp.MustCompile(`microsoft`),
}

func isBadMime(mime string) bool {
	return slices.ContainsFunc(badMimes, func(re *regexp.Regexp) bool {
		return re.MatchString(mime)
	})
}

// parse converts the retrieved data into a Clip struct.
func (c *clipboardParser) parse() (models.ClipboardEvent, error) {
	slog.Debug("parsing clipboard data")

	var event models.ClipboardEvent
	event.Time = time.Now()

	// Set MIME type
	event.MimeType = selectMime(c.mimes)

	entries := make([]models.ClipboardEntry, 0, len(c.mimes))
	for _, mime := range c.mimes {
		if isBadMime(mime) {
			continue
		}

		v, err := c.retrieveData(mime)
		if err != nil {
			return event, err
		}
		if len(v) == 0 {
			continue
		}

		hash := models.NewHash(v)

		var entry models.ClipboardEntry
		entry.Hash = hash
		entry.MimeType = mime
		entry.IsText = utf8.Valid(v)
		entry.Blob = v
		entries = append(entries, entry)
	}
	event.Entries = entries
	event.Preview = generatePreivew(entries)

	return event, nil
}
