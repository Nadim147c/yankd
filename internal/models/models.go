package models

import "time"

type Event struct {
	ID              int64     `json:"id"`
	PrimaryMimeType string    `json:"primary_mime_type"`
	Time            time.Time `json:"time"`
	Entries         []Entry   `json:"entries"`
}

type Entry struct {
	MimeType string     `json:"mime_type"`
	Hash     Hash       `json:"hash"`
	IsText   bool       `json:"is_text"`
	Text     NullString `json:"text"`
	Blob     []byte     `json:"blob"`
}
