package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Nadim147c/yankd/internal/query"
	"github.com/duckdb/duckdb-go/v2"
	"github.com/google/uuid"
)

type SearchResult struct {
	ID       uuid.UUID `json:"id"`
	Score    float64   `json:"score"`
	Time     time.Time `json:"time"`
	MimeType string    `json:"mime_type"`
	Preview  string    `json:"preview"`
}

func isDatabaseOutOfSync(err error) bool {
	v, ok := errors.AsType[*duckdb.Error](err)
	if ok && v.Type == duckdb.ErrorTypeCatalog &&
		strings.Contains(v.Msg, "Scalar Function with name match_bm25 does not exist!") {
		return true
	}
	return false
}

// BuildQuery constructs a parameterized SQL query from the given Query struct.
// It returns the query string and a slice of arguments to be passed to the driver.
func BuildQuery(q query.Query) (string, []any) {
	var simFunc string
	switch q.Flag {
	case query.Fuzzy:
		simFunc = "rapidfuzz_ratio"
	case query.Prefix:
		simFunc = "rapidfuzz_prefix_similarity"
	case query.Suffix:
		simFunc = "rapidfuzz_postfix_similarity"
	default:
		panic("unreachable")
	}

	var whereParts []string
	args := []any{q.Fuzzy}

	var typScore string
	if q.Type != "" {
		typScore = " + rapidfuzz_ratio(primary_mime_type, ?)"
		args = append(args, q.Type)
	}

	if q.After.IsValid() {
		whereParts = append(whereParts, "time > ?")
		args = append(args, q.After.StdTime())
	}
	if q.Before.IsValid() {
		whereParts = append(whereParts, "time < ?")
		args = append(args, q.Before.StdTime())
	}

	if q.Regex != "" {
		whereParts = append(whereParts, "preview SIMILAR TO ?")
		args = append(args, `.*`+q.Regex+`.*`)
	}

	if len(q.Keywords) > 0 {
		for _, kw := range q.Keywords {
			whereParts = append(whereParts, "contains(preview, ?)")
			args = append(args, kw)
		}
	}

	whereClause := "true"
	if len(whereParts) > 0 {
		whereClause = strings.Join(whereParts, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, primary_mime_type, time, preview,
          (%s(preview, ?)%s) AS score
    FROM events
		WHERE score IS NOT NULL AND %s
		ORDER BY score DESC
		LIMIT ?;
	`, simFunc, typScore, whereClause)

	args = append(args, q.Limit)
	return query, args
}

// Search runs full-text search in database and returns matched items.
// If query is empty, it returns the latest items instead.
func (db *DB) Search(ctx context.Context, q string, limit int64) ([]SearchResult, error) {
	db.mu.Lock()
	defer db.SafeUnlock()
	if limit <= 0 {
		return nil, nil
	}
	if q == "" {
		db.SafeUnlock()
		res, err := db.List(ctx, limit)
		db.mu.Lock()
		return res, err
	}

	structuredQuery, warnings := query.Parse([]byte(q))
	if len(warnings) != 0 {
		for _, warning := range warnings {
			slog.Warn("parsing query failed", "warning", warning)
		}
	}
	structuredQuery.Limit = limit

	sqlQuery, args := BuildQuery(structuredQuery)
	slog.Debug("BuildQuery", "sql-query", sqlQuery, "args", args)
	rows, err := db.sql.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		if isDatabaseOutOfSync(err) {
			slog.Info("re-trying search due to database synchronization issue")
			db.SafeUnlock()
			db.reconnect()
			return db.Search(ctx, q, limit)
		}
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
	defer db.SafeUnlock()
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
