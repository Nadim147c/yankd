package db

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var hasUpdatedIndex atomic.Bool

type SearchResult struct {
	ID       uuid.UUID `json:"id"`
	Score    float64   `json:"score"`
	Time     time.Time `json:"time"`
	MimeType string    `json:"mime_type"`
	Preview  string    `json:"preview"`
}

// Search runs full-text search in database and returns matched items.
// If query is empty, it returns the latest items instead.
func (db *DB) Search(ctx context.Context, query string, limit int64) ([]SearchResult, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if limit <= 0 {
		return nil, nil
	}

	if hasUpdatedIndex.CompareAndSwap(false, true) {
		slog.Debug("dropping the full-text search index")
		const dropFTSIndex = `
      PRAGMA drop_fts_index('events');
    `
		db.sql.ExecContext(ctx, dropFTSIndex) //nolint // No need to handle error
		slog.Debug("(re)creating the full-text search index")
		const buildFTSIndex = `
      PRAGMA create_fts_index('events', 'id', 'preview');
    `
		_, err := db.sql.ExecContext(ctx, buildFTSIndex)
		if err != nil {
			return nil, err
		}
	}

	const fts = `
    SELECT id, primary_mime_type, time, preview, score
    FROM
      (SELECT *, fts_main_events.match_bm25(id, ?) AS score FROM events) sq
    WHERE
      score IS NOT NULL
    ORDER BY
      score DESC
    Limit ?;
  `
	rows, err := db.sql.QueryContext(ctx, fts, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []SearchResult{}
	for rows.Next() {
		var i SearchResult
		if err := rows.Scan(&i.ID, &i.MimeType, &i.Time, &i.Preview, &i.Score); err != nil {
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

// List runs full-text search in database and returns matched items.
// If query is empty, it returns the latest items instead.
func (db *DB) List(ctx context.Context, limit int64) ([]SearchResult, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if limit <= 0 {
		return nil, nil
	}
	slog.Info("sqlite full-text search", "limit", limit)

	const getLastEvents = `
    SELECT id, primary_mime_type, time, preview
    FROM events
    ORDER BY time DESC
    LIMIT ?;
  `
	rows, err := db.sql.QueryContext(ctx, getLastEvents, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []SearchResult{}
	for rows.Next() {
		var i SearchResult
		if err := rows.Scan(&i.ID, &i.MimeType, &i.Time, &i.Preview); err != nil {
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
