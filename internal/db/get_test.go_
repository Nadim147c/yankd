package db

import (
	"bytes"
	"fmt"
	"slices"
	"testing"

	"github.com/Nadim147c/yankd/internal/models"
)

func TestDB_Get(t *testing.T) {
	db := setupTestDB(t)
	data := insertRandomTestData(t, db)

	for i, event := range data {
		t.Run(fmt.Sprintf("event %d", i+1), func(t *testing.T) {
			first, err := db.Get(t.Context(), event.ID)
			if err != nil {
				t.Fatalf("failed to get event: %v", err)
			}
			if first.ID != data[i].ID {
				t.Fatalf("expected %d, got %d", data[i].ID, first.ID)
			}
			if first.PrimaryMimeType != data[i].PrimaryMimeType {
				t.Fatalf("expected %s, got %s", data[i].PrimaryMimeType, first.PrimaryMimeType)
			}
			if first.Preview != data[i].Preview {
				t.Fatalf("expected %s, got %s", data[i].Preview, first.Preview)
			}
			if len(first.Entries) != len(data[i].Entries) {
				t.Fatalf("expected %d, got %d", len(data[i].Entries), len(first.Entries))
			}
			for j := range first.Entries {
				if first.Entries[j].MimeType != data[i].Entries[j].MimeType {
					t.Fatalf("expected %s, got %s", data[i].Entries[j].MimeType, first.Entries[j].MimeType)
				}
				if first.Entries[j].Hash != data[i].Entries[j].Hash {
					t.Fatalf("expected %s, got %s", data[i].Entries[j].Hash, first.Entries[j].Hash)
				}
				if first.Entries[j].IsText != data[i].Entries[j].IsText {
					t.Fatalf("expected %t, got %t", data[i].Entries[j].IsText, first.Entries[j].IsText)
				}
				if !bytes.Equal(first.Entries[j].Blob, data[i].Entries[j].Blob) {
					t.Fatalf("expected %s, got %s", data[i].Entries[j].Blob, first.Entries[j].Blob)
				}
			}
		})
	}
}

func TestDB_GetLast(t *testing.T) {
	db := setupTestDB(t)
	data := insertRandomTestData(t, db)

	n := len(data) - 1
	t.Logf("requesting last %d events", n)

	last, err := db.GetLast(t.Context(), int64(n))
	if err != nil {
		t.Fatalf("failed to get last events: %v", err)
	}
	if len(last) != n {
		t.Log(last)
		t.Fatalf("expected %d, got %d", n, len(last))
	}
}

func TestDB_GetMany(t *testing.T) {
	db := setupTestDB(t)
	data := insertRandomTestData(t, db)

	ids := []int64{data[0].ID, getLastItem(t, data).ID}
	t.Logf("requesting events with ids %v", ids)

	events, err := db.GetMany(t.Context(), ids)
	if err != nil {
		t.Fatalf("failed to get last events: %v", err)
	}
	if len(events) != len(ids) {
		t.Log(events)
		t.Fatalf("expected %d, got %d", len(ids), len(events))
	}

	if !slices.ContainsFunc(events, func(e models.ClipboardEvent) bool { return e.ID == ids[0] }) {
		t.Fatalf("expected %d to be present", ids[0])
	}

	if !slices.ContainsFunc(events, func(e models.ClipboardEvent) bool { return e.ID == ids[1] }) {
		t.Fatalf("expected %d to be present", ids[1])
	}
}
