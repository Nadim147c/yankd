package clipboard

import "time"

//go:generate go run gorm.io/cli/gorm@latest gen -i ./model.go -o ../../internal/db/binds/

// Clip is a single clipboard item
type Clip struct {
	ID       uint      `json:"id"             gorm:"uniqueIndex"`
	Time     time.Time `json:"time"`
	Text     string    `json:"text"           gorm:"uniqueIndex,class:FULLTEXT"`
	Blob     []byte    `json:"blob,omitempty"`
	Mime     string    `json:"mime"`
	Metadata string    `json:"metadata"       gorm:"uniqueIndex,class:FULLTEXT"`
	URL      string    `json:"url"            gorm:"uniqueIndex,class:FULLTEXT"`
	BlobPath string    `json:"blob_path"`
}
