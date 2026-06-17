package models

import (
	"time"

	"github.com/google/uuid"
)

type ClipboardEvent struct {
	ID       uuid.UUID        `json:"id"`
	Time     time.Time        `json:"time"`
	MimeType string           `json:"mime_type"`
	Preview  string           `json:"preview"`
	Entries  []ClipboardEntry `json:"entries"`
}

type ClipboardEntry struct {
	MimeType string `json:"mime_type"`
	Hash     Hash   `json:"hash"`
	Blob     []byte `json:"blob"`
	IsText   bool   `json:"is_text"`
}
