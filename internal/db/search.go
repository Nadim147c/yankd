package db

import (
	"context"
	"log/slog"
	"strings"

	"github.com/Nadim147c/yankd/internal/models"
)

// Search runs full-text search in database and returns matched items.
// If query is empty, it returns the latest items instead.
func (db *DB) Search(ctx context.Context, query string, limit int) ([]models.ClipboardEvent, error) {
	if limit <= 0 {
		return nil, nil
	}

	if strings.TrimSpace(query) == "" {
		return db.GetLast(ctx, int64(limit))
	}

	if stringContainsAny(query, "*", "AND", "OR", "NOT") {
		return db.fullTextSearch(ctx, query) // The query contains sqlite search oparator
	}

	fields := strings.Fields(query)
	for i, field := range fields {
		fields[i] = field + "*"
	}

	// add the *<word>* to make the serach fuzzy
	return db.fullTextSearch(ctx, strings.Join(fields, " "))
}

func (db *DB) fullTextSearch(ctx context.Context, query string) ([]models.ClipboardEvent, error) {
	slog.Info("sqlite full-text search", "query", query)
	res, err := db.queries.FullTextSearch(ctx, query)
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(res))
	_ = ids[len(res)-1] // no runtime bound check on the loop
	for i := range res {
		ids[i] = res[i].ID
	}

	return db.GetMany(ctx, ids)
}

func stringContainsAny(s string, parts ...string) bool {
	for i := range parts {
		if strings.Contains(s, parts[i]) {
			return true
		}
	}
	return false
}
