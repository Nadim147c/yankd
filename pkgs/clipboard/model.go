package clipboard

import "time"

//go:generate go run gorm.io/cli/gorm@latest gen -i ./model.go -o ../../internal/db/binds/

// Clip is a single clipboard item
type Clip struct {
	ID       uint      `json:"id"`
	Time     time.Time `json:"time"`
	Text     string    `json:"text"`
	Blob     []byte    `json:"blob,omitempty"`
	Mime     string    `json:"mime"`
	Metadata string    `json:"metadata"`
	URL      string    `json:"url,omitempty"`
	BlobPath string    `json:"blob_path,omitempty"`
}
