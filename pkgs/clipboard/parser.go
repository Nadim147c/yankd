package clipboard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	protocol "github.com/Nadim147c/yankd/internal/wlr-data-control-unstable-v1"
)

// mimeCategory returns the category of a MIME type
func mimeCategory(mimeType string) string {
	parts := strings.Split(mimeType, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return "text"
}

// isImageMime checks if MIME type is an image
func isImageMime(mimeType string) bool {
	return mimeCategory(mimeType) == "image"
}

type clipboardParser struct {
	offer         *protocol.ZwlrDataControlOfferV1
	offeredMimes  []string
	retrievedData map[string][]byte // mimeType -> data
	selectedMimes selectedMimesType
}

type selectedMimesType struct {
	primary  string // preferred image or text mime
	image    string
	text     string
	urlMime  string // chromium/x-source-url or text/x-moz-url
	metadata string // text/plain or text/html for metadata
}

// newClipboardParser creates a new parser for an offer
func newClipboardParser(
	offer *protocol.ZwlrDataControlOfferV1,
	mimes []string,
) *clipboardParser {
	slog.Debug("creating clipboard parser", "offered_mimes_count", len(mimes))
	return &clipboardParser{
		offer:         offer,
		offeredMimes:  mimes,
		retrievedData: make(map[string][]byte),
	}
}

// selectMimes determines which MIME types to retrieve
func (cp *clipboardParser) selectMimes() {
	slog.Debug(
		"selecting mime types from offered",
		"offered_count", len(cp.offeredMimes),
	)

	// Priority order for primary MIME type (prefer image over text)
	imagePriority := []string{
		"image/png",
		"image/jpeg",
		"image/webp",
		"image/gif",
	}
	textPriority := []string{
		"text/plain;charset=utf-8",
		"text/plain",
		"text/html",
	}
	urlPriority := []string{"chromium/x-source-url", "text/x-moz-url"}

	// Select image MIME
	for _, candidate := range imagePriority {
		if slices.Contains(cp.offeredMimes, candidate) {
			cp.selectedMimes.image = candidate
			cp.selectedMimes.primary = candidate
			slog.Debug("selected image mime", "mime", candidate)
			break
		}
	}

	// Select text MIME (only if no image selected)
	if cp.selectedMimes.primary == "" {
		for _, candidate := range textPriority {
			if slices.Contains(cp.offeredMimes, candidate) {
				cp.selectedMimes.text = candidate
				cp.selectedMimes.primary = candidate
				slog.Debug("selected text mime", "mime", candidate)
				break
			}
		}
	}

	// Fallback to text/plain if nothing selected
	if cp.selectedMimes.primary == "" &&
		slices.Contains(cp.offeredMimes, "text/plain") {
		cp.selectedMimes.text = "text/plain"
		cp.selectedMimes.primary = "text/plain"
		slog.Debug("fallback to text/plain mime")
	}

	// Select URL MIME
	for _, candidate := range urlPriority {
		if slices.Contains(cp.offeredMimes, candidate) {
			cp.selectedMimes.urlMime = candidate
			slog.Debug("selected url mime", "mime", candidate)
			break
		}
	}

	// Select metadata (prefer text/plain over text/html for images)
	if isImageMime(cp.selectedMimes.primary) {
		if slices.Contains(cp.offeredMimes, "text/plain") {
			cp.selectedMimes.metadata = "text/plain"
			slog.Debug("selected metadata mime for image", "mime", "text/plain")
		} else if slices.Contains(cp.offeredMimes, "text/html") {
			cp.selectedMimes.metadata = "text/html"
			slog.Debug("selected metadata mime for image", "mime", "text/html")
		}
	}

	slog.Debug(
		"mime selection complete",
		"primary", cp.selectedMimes.primary,
		"image", cp.selectedMimes.image,
		"text", cp.selectedMimes.text,
		"url", cp.selectedMimes.urlMime,
		"metadata", cp.selectedMimes.metadata,
	)
}

// getMimesToRetrieve returns the list of MIME types to fetch
func (cp *clipboardParser) getMimesToRetrieve() []string {
	var mimes []string

	if cp.selectedMimes.primary != "" {
		mimes = append(mimes, cp.selectedMimes.primary)
	}

	if cp.selectedMimes.urlMime != "" &&
		cp.selectedMimes.urlMime != cp.selectedMimes.primary {
		mimes = append(mimes, cp.selectedMimes.urlMime)
	}

	if cp.selectedMimes.metadata != "" &&
		cp.selectedMimes.metadata != cp.selectedMimes.primary {
		mimes = append(mimes, cp.selectedMimes.metadata)
	}

	slog.Debug("mime types to retrieve", "count", len(mimes), "mimes", mimes)
	return mimes
}

// retrieveData fetches data for a specific MIME type
func (cp *clipboardParser) retrieveData(mimeType string) error {
	slog.Debug("retrieving data", "mime", mimeType)

	if cp.offer == nil {
		slog.Error("offer is nil", "mime", mimeType)
		return errors.New("offer is nil")
	}

	// Create a pipe to receive data
	readFd, writeFd, err := os.Pipe()
	if err != nil {
		slog.Error("failed to create pipe", "mime", mimeType, "error", err)
		return fmt.Errorf("failed to create pipe for %s: %w", mimeType, err)
	}
	defer writeFd.Close()

	// Send receive request
	if err := cp.offer.Receive(mimeType, uintptr(writeFd.Fd())); err != nil {
		readFd.Close()
		slog.Error("receive request failed", "mime", mimeType, "error", err)
		return fmt.Errorf("receive request failed for %s: %w", mimeType, err)
	}

	// Close write end in this process
	writeFd.Close()

	// Read data from the read end
	data, err := io.ReadAll(readFd)
	readFd.Close()

	if err != nil {
		slog.Error("failed to read data", "mime", mimeType, "error", err)
		return fmt.Errorf("failed to read data for %s: %w", mimeType, err)
	}

	cp.retrievedData[mimeType] = data
	slog.Debug(
		"data retrieved successfully",
		"mime", mimeType,
		"size_bytes", len(data),
	)
	return nil
}

// RetrieveAll fetches all selected MIME types
func (cp *clipboardParser) RetrieveAll() error {
	slog.Debug("retrieving all selected mime types")

	cp.selectMimes()
	mimes := cp.getMimesToRetrieve()

	if len(mimes) == 0 {
		slog.Error("no suitable mime types to retrieve")
		return errors.New("no suitable MIME types to retrieve")
	}

	for _, mime := range mimes {
		if err := cp.retrieveData(mime); err != nil {
			slog.Warn(
				"failed to retrieve mime type, continuing with others",
				"mime",
				mime,
				"error",
				err,
			)
			// Continue with other MIME types
		}
	}

	slog.Debug(
		"retrieve all completed",
		"retrieved_count",
		len(cp.retrievedData),
		"requested_count",
		len(mimes),
	)
	return nil
}

// Parse converts the retrieved data into a Clip struct
func (cp *clipboardParser) Parse() (Clip, error) {
	slog.Debug("parsing clipboard data")

	clip := Clip{Time: time.Now()}

	if err := cp.RetrieveAll(); err != nil {
		slog.Error("failed to retrieve all mime types", "error", err)
		return clip, err
	}

	// Set MIME type
	clip.Mime = cp.selectedMimes.primary
	if clip.Mime == "" {
		clip.Mime = "text/plain"
		slog.Debug("mime type defaulted to text/plain")
	}

	// Handle image data
	if isImageMime(cp.selectedMimes.primary) {
		slog.Debug("parsing image data", "mime", cp.selectedMimes.primary)

		if data, ok := cp.retrievedData[cp.selectedMimes.primary]; ok {
			clip.Blob = data
			slog.Debug("image blob set", "size_bytes", len(data))
		}

		// Get metadata for image
		if cp.selectedMimes.metadata != "" {
			if data, ok := cp.retrievedData[cp.selectedMimes.metadata]; ok {
				clip.Metadata = string(data)
				slog.Debug("image metadata set", "size_bytes", len(data))
			}
		}
	} else {
		// Handle text data
		slog.Debug("parsing text data", "mime", cp.selectedMimes.primary)

		if data, ok := cp.retrievedData[cp.selectedMimes.primary]; ok {
			clip.Text = string(bytes.TrimSpace(data))
			slog.Debug("text data set", "length", len(clip.Text))
		}
	}

	// Get URL
	if cp.selectedMimes.urlMime != "" {
		if data, ok := cp.retrievedData[cp.selectedMimes.urlMime]; ok {
			// URL might have trailing newlines or extra whitespace
			clip.URL = string(bytes.TrimSpace(data))
			slog.Debug("url set", "length", len(clip.URL))
		}
	}

	slog.Info(
		"clipboard parsed successfully",
		"id", cp.offer.Id(),
		"mime", clip.Mime,
		"has_blob", len(clip.Blob) > 0,
		"has_text", len(clip.Text) > 0,
		"has_url", len(clip.URL) > 0,
		"has_metadata", len(clip.Metadata) > 0,
	)

	return clip, nil
}
