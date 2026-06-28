package db

import (
	"testing"
)

func TestDB_Search(t *testing.T) {
	db := setupTestDB(t)
	data := insertRandomTestData(t, db)

	count := len(data)
	t.Run("No query", func(t *testing.T) {
		res, err := db.Search(t.Context(), "", int64(count))
		if err != nil {
			t.Fatalf("failed to search: %v", err)
		}
		if len(res) != count {
			t.Fatalf("expected %d, got %d", count, len(res))
		}
	})

	t.Run("first item preview", func(t *testing.T) {
		res, err := db.Search(t.Context(), data[0].Preview, int64(count))
		if err != nil {
			t.Fatalf("failed to search: %v", err)
		}

		if res[0].ID != data[0].ID {
			t.Fatalf("expected %d, got %d", data[0].ID, res[0].ID)
		}
	})
}
