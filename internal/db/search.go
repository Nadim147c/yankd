package db

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/Nadim147c/yankd/internal/db/sqlc"
	fzfAlgo "github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

// Search runs full-text search in database and returns matched items.
// If query is empty, it returns the latest items instead.
func (db *DB) Search(ctx context.Context, query string, limit int64) ([]sqlc.GetEventsPreviewAndIDRow, error) {
	if limit <= 0 {
		return nil, nil
	}

	if strings.TrimSpace(query) == "" {
		return db.queries.GetEventsPreviewAndID(ctx, limit)
	}

	slog.Info("sqlite full-text search", "query", query)
	previews, err := db.queries.GetEventsPreviewAndID(ctx, 10000000)
	if err != nil || len(previews) == 0 {
		return nil, err
	}

	pattern := []rune(query)
	slab := util.MakeSlab(100*1024, 2048) // Pre-allocate memory for efficiency

	type scoredEvent struct {
		score   int
		preview sqlc.GetEventsPreviewAndIDRow
	}

	var matches []scoredEvent

	// 2. Iterate and Match
	for _, event := range previews {
		// Convert string to fzf's expected Chars type
		input := util.ToChars([]byte(event.Preview))

		// Run the algorithm
		// Parameters: caseSensitive, normalize, forward, input, pattern, withPos, slab
		res, _ := fzfAlgo.FuzzyMatchV2(false, true, true, &input, pattern, false, slab)

		if res.Score > 0 {
			matches = append(matches, scoredEvent{res.Score, event})
		}
	}

	slices.SortStableFunc(matches, func(a, b scoredEvent) int {
		return b.score - a.score
	})

	size := min(len(matches), int(limit))

	for i, m := range matches[:size] {
		previews[i] = m.preview
	}

	// use already allocated preview slice
	return previews[:size], nil
}

// List runs full-text search in database and returns matched items.
// If query is empty, it returns the latest items instead.
func (db *DB) List(ctx context.Context, limit int64) ([]sqlc.GetEventsPreviewAndIDRow, error) {
	if limit <= 0 {
		return nil, nil
	}
	slog.Info("sqlite full-text search", "limit", limit)
	return db.queries.GetEventsPreviewAndID(ctx, limit)
}
