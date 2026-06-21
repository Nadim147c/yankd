package db

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/Nadim147c/yankd/internal/models"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()

	internalTestModeDoNotUse = true
	db, err := CreateDB()
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("failed to close test database: %v", err)
		}
	})

	return db
}

func getTestTextEntry(t *testing.T, text string) models.ClipboardEntry {
	t.Helper()
	blob := []byte(text)
	return models.ClipboardEntry{
		IsText:   true,
		MimeType: "text/plain",
		Hash:     models.NewHash(blob),
		Blob:     blob,
	}
}

func getTestBinaryEntry(t *testing.T, mimeType string, size int) models.ClipboardEntry {
	t.Helper()
	blob := make([]byte, size)
	rand.Read(blob) //nolint // rand.Read never fails (panics on failure)
	return models.ClipboardEntry{
		IsText:   false,
		MimeType: mimeType,
		Hash:     models.NewHash(blob),
		Blob:     blob,
	}
}

func insertRandomTestData(t *testing.T, db *DB) []*models.ClipboardEvent {
	t.Helper()

	now := time.Now()
	events := []*models.ClipboardEvent{
		{
			MimeType: "image/png",
			Time:     now,
			Entries: []models.ClipboardEntry{
				getTestTextEntry(t, "hello"),
				getTestTextEntry(t, "world"),
				getTestBinaryEntry(t, "image/png", 1024),
			},
			Preview: "hello world",
		},
		{
			MimeType: "text/plain",
			Time:     now,
			Entries: []models.ClipboardEntry{
				getTestTextEntry(t, "lorem ipsum dolor sit amet"),
			},
			Preview: "lorem ipsum dolor sit amet",
		},
		{
			MimeType: "image/jpeg",
			Time:     now,
			Entries: []models.ClipboardEntry{
				getTestTextEntry(t, "second"),
				getTestTextEntry(t, "clipboard"),
				getTestBinaryEntry(t, "image/jpeg", 1024),
			},
			Preview: "second clipboard",
		},
	}

	for _, event := range events {
		if err := db.Insert(t.Context(), event); err != nil {
			t.Fatalf("failed to insert event: %v", err)
		}
	}
	return events
}

func getLastItem(t *testing.T, events []*models.ClipboardEvent) *models.ClipboardEvent {
	t.Helper()
	if len(events) == 0 {
		t.Fatalf("events is empty")
	}
	return events[len(events)-1]
}
