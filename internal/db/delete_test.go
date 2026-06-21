package db

import (
	"testing"

	"github.com/google/uuid"
)

func TestDB_Delete(t *testing.T) {
	t.Run("delete all item except the last one", func(t *testing.T) {
		db := setupTestDB(t)
		currentEvents := insertRandomTestData(t, db)
		idsToDelete := make([]uuid.UUID, len(currentEvents)-1)
		for i, event := range currentEvents[:len(currentEvents)-1] {
			idsToDelete[i] = event.ID
		}

		_, err := db.Delete(t.Context(), idsToDelete...)
		if err != nil {
			t.Fatalf("failed to delete events: %v", err)
		}
		remainingEvents, err := db.List(t.Context(), 10)
		if err != nil {
			t.Fatalf("failed to get last events: %v", err)
		}
		if len(remainingEvents) != 1 {
			t.Fatalf("expected 1 remaining event, got %d", len(remainingEvents))
		}
	})

	t.Run("delete non-existing id", func(t *testing.T) {
		db := setupTestDB(t)
		data := insertRandomTestData(t, db)
		idsToDelete := []uuid.UUID{uuid.New()}
		deletedCount, err := db.Delete(t.Context(), idsToDelete...)
		if err != nil {
			t.Fatalf("failed to delete events: %v", err)
		}
		if deletedCount != 0 {
			t.Fatalf("expected 0 deleted events, got %d", deletedCount)
		}

		// TODO: add db.CountEvent
		eventCount, err := db.List(t.Context(), int64(len(data)))
		if err != nil {
			t.Fatalf("failed to get last events: %v", err)
		}
		if len(eventCount) != len(data) {
			t.Fatalf("expected %d remaining events, got %d", len(data), len(eventCount))
		}
	})

	t.Run("delete all events", func(t *testing.T) {
		db := setupTestDB(t)
		data := insertRandomTestData(t, db)
		_, err := db.Wipe(t.Context())
		if err != nil {
			t.Fatalf("failed to wipe events: %v", err)
		}

		eventCount, err := db.List(t.Context(), int64(len(data)))
		if err != nil {
			t.Fatalf("failed to get last events: %v", err)
		}
		if len(eventCount) != 0 {
			t.Fatalf("expected 0 remaining events, got %d", len(eventCount))
		}
	})
}

func TestDB_Wipe(t *testing.T) {
	db := setupTestDB(t)
	data := insertRandomTestData(t, db)

	_, err := db.Wipe(t.Context())
	if err != nil {
		t.Fatalf("failed to wipe events: %v", err)
	}
	eventCount, err := db.List(t.Context(), int64(len(data)))
	if err != nil {
		t.Fatalf("failed to get last events: %v", err)
	}
	if len(eventCount) != 0 {
		t.Fatalf("expected 0 remaining events, got %d", len(eventCount))
	}
}
