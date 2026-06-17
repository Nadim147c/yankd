package db

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Nadim147c/yankd/internal/models"
	"github.com/google/uuid"
)

type ClipboardEntryWithEventID struct {
	EventID uuid.UUID
	models.ClipboardEntry
}

func convertUUIDsToString(ids []uuid.UUID) string {
	if len(ids) == 0 {
		return ""
	}
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "'%s'::UUID", ids[0])
	for _, id := range ids[1:] {
		buf.WriteByte(',')
		fmt.Fprintf(buf, "'%s'::UUID", id)
	}
	return buf.String()
}

func (db *DB) getEntries(ctx context.Context, ids ...uuid.UUID) ([]ClipboardEntryWithEventID, error) {
	const getEntries = `
    SELECT entries.event_id, entries.mime_type, contents.hash, contents.is_text, contents.blob FROM entries
    JOIN contents ON contents.id = entries.content_id
    WHERE entries.event_id IN [@UUID_ARRAY@]
  `
	query := strings.ReplaceAll(getEntries, "@UUID_ARRAY@", convertUUIDsToString(ids))
	rows, err := db.sql.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ClipboardEntryWithEventID{}
	for rows.Next() {
		var i ClipboardEntryWithEventID
		if err := rows.Scan(&i.EventID, &i.MimeType, &i.Hash, &i.IsText, &i.Blob); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// Get return a single clipboard item from history.
func (db *DB) Get(ctx context.Context, id uuid.UUID) (m models.ClipboardEvent, err error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	const getEvent = `
    SELECT id, primary_mime_type, time, preview FROM events
    WHERE id = ?
  `
	row := db.sql.QueryRowContext(ctx, getEvent, id)
	err = row.Scan(
		&m.ID,
		&m.MimeType,
		&m.Time,
		&m.Preview,
	)
	if err != nil {
		return m, fmt.Errorf("failed to get event: %w", err)
	}

	entries, err := db.getEntries(ctx, id)
	if err != nil {
		return m, fmt.Errorf("failed to get entries: %w", err)
	}
	m.Entries = make([]models.ClipboardEntry, len(entries))
	_ = m.Entries[:len(entries)]
	for i := range entries {
		m.Entries[i] = entries[i].ClipboardEntry
	}
	return m, nil
}

// GetMany return a single clipboard item from history.
func (db *DB) GetMany(ctx context.Context, ids ...uuid.UUID) ([]models.ClipboardEvent, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	const getEvents = `
    SELECT id, primary_mime_type, time, preview FROM events
    WHERE id IN [@UUID_ARRAY@]
  `
	query := strings.ReplaceAll(getEvents, "@UUID_ARRAY@", convertUUIDsToString(ids))
	rows, err := db.sql.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []models.ClipboardEvent{}
	for rows.Next() {
		var e models.ClipboardEvent
		if err := rows.Scan(&e.ID, &e.MimeType, &e.Time, &e.Preview); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	entries, err := db.getEntries(ctx, ids...)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// group entries by event_id
	entryMap := make(map[uuid.UUID][]models.ClipboardEntry)
	for en := range slices.Values(entries) {
		entryMap[en.EventID] = append(entryMap[en.EventID], en.ClipboardEntry)
	}

	eventMap := make(map[uuid.UUID]models.ClipboardEvent)
	for _, ev := range events {
		eventMap[ev.ID] = ev
	}

	result := make([]models.ClipboardEvent, 0, len(ids))
	for _, id := range ids {
		if ev, exists := eventMap[id]; exists {
			ev.Entries = entryMap[ev.ID]
			result = append(result, ev)
		}
	}

	return result, nil
}
