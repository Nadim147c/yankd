package clipboard

import (
	"encoding/pem"
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

var ErrContainsSecrets = errors.New("clipboard content contains secrets")

func containsPEMSecret(b []byte) bool {
	for {
		var block *pem.Block
		block, b = pem.Decode(b)
		if block == nil {
			break
		}
		typeUpper := strings.ToUpper(block.Type)
		// Explicitly allow known non-secret public key variants
		if strings.Contains(typeUpper, "PUBLIC KEY") {
			continue
		}
		return true
	}

	return false
}

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

// mimeReadTimeout is how long retrieveData waits for the next chunk of a MIME
// type. A source may advertise a target and never answer it, which otherwise
// blocks the read forever and stalls the whole watcher until the next
// selection change. It is an idle timeout, reset after every successful read,
// so a large payload that keeps streaming is never cut off.
const mimeReadTimeout = time.Second

// readWithIdleTimeout reads r until EOF and returns the data read. It gives up
// once idleTimeout passes without new data; the caller must discard the data
// when err is non-nil.
func readWithIdleTimeout(r *os.File, idleTimeout time.Duration) ([]byte, error) {
	var data []byte
	buf := make([]byte, 32*1024)

	for {
		if err := r.SetReadDeadline(time.Now().Add(idleTimeout)); err != nil {
			return data, err
		}

		n, err := r.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return data, nil
			}
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return data, fmt.Errorf(
					"idle timeout after %s with %d bytes read: %w",
					idleTimeout, len(data), err,
				)
			}
			return data, err
		}
	}
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
	data, err := readWithIdleTimeout(reader, mimeReadTimeout)
	reader.Close() //nolint

	if err != nil {
		slog.Error("failed to read data", "mime", mimeType, "error", err)
		return nil, fmt.Errorf("failed to read data for %s: %w", mimeType, err)
	}

	return data, nil
}

// SelectMimeType selects best mime for current clipboard item. Prefer image/* mimes
// with valid file extensions. Fallback: text/plain + ".txt".
func SelectMimeType(m []string) (mime string) {
	// First pass: look for image/*
	for _, mt := range m {
		mtype, _, _ := mimepkg.ParseMediaType(mt)
		if strings.HasPrefix(mtype, "image/") {
			return mt
		}
		if strings.HasPrefix(mtype, "video/") {
			return mt
		}
		if strings.HasPrefix(mtype, "audio/") {
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

// controlTargets are X11 ICCCM meta targets, not clipboard content: TARGETS
// answers with a list of names, TIMESTAMP with a number, and SAVE_TARGETS only
// starts the clipboard-manager save handshake. GTK advertises SAVE_TARGETS but
// never writes to the fd or closes it, which blocks retrieveData on the pipe
// until the source dies. The uppercase X11 data targets (STRING, UTF8_STRING,
// TEXT, COMPOUND_TEXT) are deliberately absent: they carry real text.
var controlTargets = map[string]bool{
	"SAVE_TARGETS":     true,
	"TARGETS":          true,
	"TIMESTAMP":        true,
	"MULTIPLE":         true,
	"DELETE":           true,
	"INSERT_SELECTION": true,
	"INSERT_PROPERTY":  true,
	"LENGTH":           true,
}

// parse converts the retrieved data into a Clip struct.
func (c *clipboardParser) parse() (models.ClipboardEvent, error) {
	slog.Debug("parsing clipboard data")

	var event models.ClipboardEvent
	event.Time = time.Now()

	// Set MIME type
	event.MimeType = SelectMimeType(c.mimes)

	entries := make([]models.ClipboardEntry, 0, len(c.mimes))
	for _, mime := range c.mimes {
		if mime == "x-kde-passwordManagerHint" {
			return models.ClipboardEvent{}, ErrContainsSecrets
		}

		if isBadMime(mime) || controlTargets[mime] {
			continue
		}

		v, err := c.retrieveData(mime)
		if err != nil {
			// A single unreadable MIME type must not drop the whole event.
			// retrieveData already logged the error.
			continue
		}
		if len(v) == 0 {
			continue
		}

		if containsPEMSecret(v) {
			return models.ClipboardEvent{}, ErrContainsSecrets
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
	event.Preview = GeneratePreivew(entries)

	return event, nil
}
