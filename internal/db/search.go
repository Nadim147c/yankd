package db

import (
	"context"
	"log/slog"
	"slices"
	"strings"

	"github.com/Nadim147c/yankd/internal/models"
	fzfAlgo "github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

// Search runs full-text search in database and returns matched items.
// If query is empty, it returns the latest items instead.
func (db *DB) Search(ctx context.Context, query string, limit int64) ([]models.ClipboardEvent, error) {
	if limit <= 0 {
		return nil, nil
	}

	if strings.TrimSpace(query) == "" {
		return db.GetLast(ctx, limit)
	}

	slog.Info("sqlite full-text search", "query", query)
	previews, err := db.queries.GetEventsPreviewAndID(ctx, limit)
	if err != nil || len(previews) == 0 {
		return nil, err
	}

	pattern := []rune(query)
	slab := util.MakeSlab(100*1024, 2048) // Pre-allocate memory for efficiency

	type scoredEvent struct {
		id    int64
		score int
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
			matches = append(matches, scoredEvent{
				id:    event.ID,
				score: res.Score,
			})
		}
	}

	slices.SortStableFunc(matches, func(a, b scoredEvent) int { return b.score - a.score })

	resultIDs := make([]int64, len(matches))
	for i, m := range matches {
		resultIDs[i] = m.id
	}

	return db.GetMany(ctx, resultIDs)
}
