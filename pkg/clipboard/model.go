package clipboard

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"

	"github.com/cespare/xxhash/v2"
)

//go:generate gorm gen -i ./model.go -o ../../internal/db/binds/

// Hash is a uint64 wrapper
type Hash uint64

var (
	_ fmt.Stringer  = (*Hash)(nil)
	_ sql.Scanner   = (*Hash)(nil)
	_ driver.Valuer = (*Hash)(nil)
)

// String convert hash to String
func (h Hash) String() string {
	return strconv.FormatUint(uint64(h), 16)
}

// Value import driver.Valuer
func (h Hash) Value() (driver.Value, error) {
	return h.String(), nil
}

// Scan implements sql.Scaner
func (h *Hash) Scan(value any) error {
	if value == nil {
		*h = 0
		return nil
	}

	switch v := value.(type) {
	case string: // TEXT column
		u, err := strconv.ParseUint(v, 16, 64)
		if err != nil {
			return fmt.Errorf("failed to scan Hash(string): %w", err)
		}
		*h = Hash(u)
		return nil

	case []byte: // SQLite may return TEXT as BLOB
		u, err := strconv.ParseUint(string(v), 16, 64)
		if err != nil {
			return fmt.Errorf("failed to scan Hash([]byte): %w", err)
		}
		*h = Hash(u)
		return nil

	default: // should never happen if column is TEXT, but safe guard anyway
		return fmt.Errorf("unsupported type scanned for Hash: %T", value)
	}
}

// HashClip returns uint64 hash for clip content
func HashClip(clip Clip) Hash {
	w := xxhash.New()
	w.WriteString(clip.Mime)
	w.WriteString(clip.Text)
	w.WriteString(clip.Metadata)
	w.WriteString(clip.URL)
	w.Write(clip.Blob)
	return Hash(w.Sum64())
}

// Clip is a single clipboard item
type Clip struct {
	ID       uint      `json:"id"`
	Time     time.Time `json:"time"`
	Hash     Hash      `json:"hash"                gorm:"index:,unique,length:16"`
	Text     string    `json:"text"`
	Mime     string    `json:"mime"`
	Metadata string    `json:"metadata"`
	URL      string    `json:"url,omitempty"`
	Blob     []byte    `json:"blob,omitempty"`
	BlobPath string    `json:"blob_path,omitempty"`
	BlobHash Hash      `json:"blob_hash,omitempty" gorm:"index:,length:16"`
}
