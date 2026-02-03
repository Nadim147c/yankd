package db

import (
	"context"
	"strings"

	"github.com/Nadim147c/yankd/internal/models"
)

// Search runs full-text search in database and returns matched items.
// If query is empty, it returns the latest items instead.
func (db *DB) Search(ctx context.Context, query string, limit int) ([]models.Event, error) {
	if limit <= 0 {
		return nil, nil
	}

	if strings.TrimSpace(query) == "" {
		return db.GetLast(ctx, int64(limit))
	}

	res, err := db.queries.FullTextSearch(ctx, query)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(res))
	for i := range res {
		ids[i] = res[i].ID
	}

	return db.GetMany(ctx, ids)
}
