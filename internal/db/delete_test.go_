package db

import (
	"math"
	"testing"
)

func TestDB_Delete(t *testing.T) {
	t.Run("delete all item except the last one", func(t *testing.T) {
		db := setupTestDB(t)
		currentEvents := insertRandomTestData(t, db)
		idsToDelete := make([]int64, len(currentEvents)-1)
		for i, event := range currentEvents[:len(currentEvents)-1] {
			idsToDelete[i] = event.ID
		}

		deletedCount, err := db.Delete(t.Context(), idsToDelete)
		if err != nil {
			t.Fatalf("failed to delete events: %v", err)
		}
		if deletedCount != 2 {
			t.Fatalf("expected 2 deleted events, got %d", deletedCount)
		}

		remainingEvents, err := db.GetLast(t.Context(), getLastItem(t, currentEvents).ID)
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
		idsToDelete := []int64{math.MaxInt64}
		deletedCount, err := db.Delete(t.Context(), idsToDelete)
		if err != nil {
			t.Fatalf("failed to delete events: %v", err)
		}
		if deletedCount != 0 {
			t.Fatalf("expected 0 deleted events, got %d", deletedCount)
		}

		// TODO: add db.CountEvent
		eventCount, err := db.GetLast(t.Context(), int64(len(data)))
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
		deletedCount, err := db.Wipe(t.Context())
		if err != nil {
			t.Fatalf("failed to wipe events: %v", err)
		}
		if deletedCount != int64(len(data)) {
			t.Fatalf("expected %d deleted events, got %d", len(data), deletedCount)
		}

		// TODO: add db.CountEvent
		eventCount, err := db.GetLast(t.Context(), int64(len(data)))
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

	n, err := db.Wipe(t.Context())
	if err != nil {
		t.Fatalf("failed to wipe events: %v", err)
	}
	if n != int64(len(data)) {
		t.Fatalf("expected %d deleted events, got %d", len(data), n)
	}
}
