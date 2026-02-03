package db

import (
	"context"
	"fmt"
	"slices"

	"github.com/Nadim147c/yankd/internal/db/sqlc"
	"github.com/Nadim147c/yankd/internal/models"
)

func ModelConvert(e sqlc.Event, entries []sqlc.GetEntriesRow) models.Event {
	clipEntries := make([]models.Entry, len(entries))
	for i, e := range entries {
		clipEntries[i] = models.Entry{
			MimeType: e.MimeType,
			Hash:     e.Hash,
			IsText:   e.IsText,
			Text:     e.Text,
			Blob:     e.Blob,
		}
	}
	return models.Event{
		ID:              e.ID,
		Time:            e.Time,
		PrimaryMimeType: e.PrimaryMimeType,
		Entries:         clipEntries,
	}
}

// Get return a single clipboard item from history.
func (db *DB) Get(ctx context.Context, id int64) (models.Event, error) {
	event, err := db.queries.GetEvent(ctx, id)
	if err != nil {
		return models.Event{}, fmt.Errorf("failed to get event: %w", err)
	}
	entries, err := db.queries.GetEntries(ctx, []int64{id})
	if err != nil {
		return models.Event{}, fmt.Errorf("failed to get entries: %w", err)
	}
	return ModelConvert(event, entries), nil
}

// Get return a single clipboard item from history.
func (db *DB) GetLast(ctx context.Context, limit int64) ([]models.Event, error) {
	events, err := db.queries.GetLastEvents(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	ids := make([]int64, len(events))
	for event := range slices.Values(events) {
		ids = append(ids, event.ID)
	}

	entries, err := db.queries.GetEntries(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// group entries by event_id
	entryMap := make(map[int64][]sqlc.GetEntriesRow)
	for en := range slices.Values(entries) {
		entryMap[en.EventID] = append(entryMap[en.EventID], en)
	}

	result := make([]models.Event, len(events))
	for i, ev := range events {
		result[i] = ModelConvert(ev, entryMap[ev.ID])
	}

	return result, nil
}

// GetMany return a single clipboard item from history.
func (db *DB) GetMany(ctx context.Context, ids []int64) ([]models.Event, error) {
	events, err := db.queries.GetEvents(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	entries, err := db.queries.GetEntries(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// group entries by event_id
	entryMap := make(map[int64][]sqlc.GetEntriesRow)
	for en := range slices.Values(entries) {
		entryMap[en.EventID] = append(entryMap[en.EventID], en)
	}

	result := make([]models.Event, len(events))
	for i, ev := range events {
		result[i] = ModelConvert(ev, entryMap[ev.ID])
	}

	return result, nil
}
