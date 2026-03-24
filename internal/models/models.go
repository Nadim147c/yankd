package models

import "time"

type ClipboardEvent struct {
	ID              int64            `json:"id"`
	PrimaryMimeType string           `json:"primary_mime_type"`
	Time            time.Time        `json:"time"`
	Entries         []ClipboardEntry `json:"entries"`
	Preview         string           `json:"preview"`
}

type ClipboardEntry struct {
	MimeType string `json:"mime_type"`
	Hash     Hash   `json:"hash"`
	IsText   bool   `json:"is_text"`
	Blob     []byte `json:"blob"`
}
