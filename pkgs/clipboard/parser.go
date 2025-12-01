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

type ClipboardParser struct {
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

// NewClipboardParser creates a new parser for an offer
func NewClipboardParser(offer *protocol.ZwlrDataControlOfferV1, mimes []string) *ClipboardParser {
	return &ClipboardParser{
		offer:         offer,
		offeredMimes:  mimes,
		retrievedData: make(map[string][]byte),
	}
}

// selectMimes determines which MIME types to retrieve
func (cp *ClipboardParser) selectMimes() {
	// Priority order for primary MIME type (prefer image over text)
	imagePriority := []string{"image/png", "image/jpeg", "image/webp", "image/gif"}
	textPriority := []string{"text/plain;charset=utf-8", "text/plain", "text/html"}
	urlPriority := []string{"chromium/x-source-url", "text/x-moz-url"}

	// Select image MIME
	for _, candidate := range imagePriority {
		if slices.Contains(cp.offeredMimes, candidate) {
			cp.selectedMimes.image = candidate
			cp.selectedMimes.primary = candidate
			break
		}
	}

	// Select text MIME (only if no image selected)
	if cp.selectedMimes.primary == "" {
		for _, candidate := range textPriority {
			if slices.Contains(cp.offeredMimes, candidate) {
				cp.selectedMimes.text = candidate
				cp.selectedMimes.primary = candidate
				break
			}
		}
	}

	// Fallback to text/plain if nothing selected
	if cp.selectedMimes.primary == "" && slices.Contains(cp.offeredMimes, "text/plain") {
		cp.selectedMimes.text = "text/plain"
		cp.selectedMimes.primary = "text/plain"
	}

	// Select URL MIME
	for _, candidate := range urlPriority {
		if slices.Contains(cp.offeredMimes, candidate) {
			cp.selectedMimes.urlMime = candidate
			break
		}
	}

	// Select metadata (prefer text/plain over text/html for images)
	if isImageMime(cp.selectedMimes.primary) {
		if slices.Contains(cp.offeredMimes, "text/plain") {
			cp.selectedMimes.metadata = "text/plain"
		} else if slices.Contains(cp.offeredMimes, "text/html") {
			cp.selectedMimes.metadata = "text/html"
		}
	}
}

// getMimesToRetrieve returns the list of MIME types to fetch
func (cp *ClipboardParser) getMimesToRetrieve() []string {
	var mimes []string

	if cp.selectedMimes.primary != "" {
		mimes = append(mimes, cp.selectedMimes.primary)
	}
	if cp.selectedMimes.urlMime != "" && cp.selectedMimes.urlMime != cp.selectedMimes.primary {
		mimes = append(mimes, cp.selectedMimes.urlMime)
	}
	if cp.selectedMimes.metadata != "" && cp.selectedMimes.metadata != cp.selectedMimes.primary {
		mimes = append(mimes, cp.selectedMimes.metadata)
	}

	return mimes
}

// retrieveData fetches data for a specific MIME type
func (cp *ClipboardParser) retrieveData(mimeType string) error {
	if cp.offer == nil {
		return errors.New("offer is nil")
	}

	// Create a pipe to receive data
	readFd, writeFd, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe for %s: %w", mimeType, err)
	}
	defer writeFd.Close()

	// Send receive request
	if err := cp.offer.Receive(mimeType, uintptr(writeFd.Fd())); err != nil {
		readFd.Close()
		return fmt.Errorf("receive request failed for %s: %w", mimeType, err)
	}

	// Close write end in this process
	writeFd.Close()

	// Read data from the read end
	data, err := io.ReadAll(readFd)
	readFd.Close()

	if err != nil {
		return fmt.Errorf("failed to read data for %s: %w", mimeType, err)
	}

	cp.retrievedData[mimeType] = data
	slog.Debug("Retrieved", "mime", mimeType, "size", len(data))
	return nil
}

// RetrieveAll fetches all selected MIME types
func (cp *ClipboardParser) RetrieveAll() error {
	cp.selectMimes()
	mimes := cp.getMimesToRetrieve()

	if len(mimes) == 0 {
		return errors.New("no suitable MIME types to retrieve")
	}

	for _, mime := range mimes {
		if err := cp.retrieveData(mime); err != nil {
			fmt.Printf("[Warning] %v\n", err)
			// Continue with other MIME types
		}
	}

	return nil
}

// Parse converts the retrieved data into a Clip struct
func (cp *ClipboardParser) Parse() (Clip, error) {
	clip := Clip{Time: time.Now()}

	if err := cp.RetrieveAll(); err != nil {
		return clip, err
	}

	// Set MIME type
	clip.Mime = cp.selectedMimes.primary
	if clip.Mime == "" {
		clip.Mime = "text/plain"
	}

	// Handle image data
	if isImageMime(cp.selectedMimes.primary) {
		if data, ok := cp.retrievedData[cp.selectedMimes.primary]; ok {
			clip.Blob = data
		}

		// Get metadata for image
		if cp.selectedMimes.metadata != "" {
			if data, ok := cp.retrievedData[cp.selectedMimes.metadata]; ok {
				clip.Metadata = string(data)
			}
		}
	} else {
		// Handle text data
		if data, ok := cp.retrievedData[cp.selectedMimes.primary]; ok {
			clip.Text = string(bytes.TrimSpace(data))
		}
	}

	// Get URL
	if cp.selectedMimes.urlMime != "" {
		if data, ok := cp.retrievedData[cp.selectedMimes.urlMime]; ok {
			// URL might have trailing newlines or extra whitespace
			clip.URL = string(bytes.TrimSpace(data))
		}
	}

	return clip, nil
}
