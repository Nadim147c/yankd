package clipboard

import "time"

type Clip struct {
	ID       uint      `json:"id"`
	Time     time.Time `json:"time"`
	Text     string    `json:"text"`
	Blob     []byte    `json:"blob,omitempty"`
	Mime     string    `json:"mime"`
	Metadata string    `json:"metadata"`
	URL      string    `json:"url"`
	BlobPath string    `json:"blob_path"`
}
