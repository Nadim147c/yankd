package db

import (
	"testing"
	"time"

	"github.com/Nadim147c/yankd/internal/models"
)

func TestDB_Insert(t *testing.T) {
	db := setupTestDB(t)

	entry := getTestBinaryEntry(t, "image/png", 1024)
	event := models.ClipboardEvent{
		MimeType: "image/png",
		Time:     time.Now(),
		Entries: []models.ClipboardEntry{
			getTestTextEntry(t, "hello"),
			getTestTextEntry(t, "world"),
			entry,
			entry,
		},
		Preview: "hello world",
	}

	if err := db.Insert(t.Context(), &event); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	var count int
	if err := db.sql.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM contents").Scan(&count); err != nil {
		t.Fatalf("failed to count contents: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected %d, got %d", 3, count)
	}

	if err := db.sql.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM entries").Scan(&count); err != nil {
		t.Fatalf("failed to count entries: %v", err)
	}
	if count != 4 {
		t.Fatalf("expected %d, got %d", 4, count)
	}

	if err := db.sql.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM events").Scan(&count); err != nil {
		t.Fatalf("failed to count events: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected %d, got %d", 1, count)
	}
}
