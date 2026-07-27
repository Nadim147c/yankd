package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/Nadim147c/yankd/internal/query"
	"github.com/duckdb/duckdb-go/v2"
	"github.com/google/uuid"
)

type SearchResult struct {
	ID       uuid.UUID `json:"id"`
	Time     time.Time `json:"time"`
	MimeType string    `json:"mime_type"`
	Preview  string    `json:"preview"`
	Score    float64   `json:"score"`
}

// isDatabaseOutOfSync is a hack!
// This is what you get for using unreliable database.
func isDatabaseOutOfSync(err error) bool {
	v, ok := errors.AsType[*duckdb.Error](err)
	return ok && v.Type == duckdb.ErrorTypeCatalog &&
		strings.Contains(v.Msg, "does not exist!")
}

// BuildQuery constructs a parameterized SQL query from the given Query struct.
// It returns the query string and a slice of arguments to be passed to the driver.
func BuildQuery(q query.Query) (string, []any) {
	var whereParts []string
	args := []any{}
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

	if q.Type != "" {
		whereParts = append(whereParts, "type_score > 70")
	}

	whereClause := "true"
	if len(whereParts) > 0 {
		whereClause = strings.Join(whereParts, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, primary_mime_type, time, preview,
       rapidfuzz_ratio              (preview, ?) as ratio_score,
       rapidfuzz_osa_similarity     (preview, ?) as osa_score,
       rapidfuzz_lcs_seq_similarity (preview, ?) as lcs_score,
       rapidfuzz_prefix_similarity  (preview, ?) as prefix_score,
       rapidfuzz_postfix_similarity (preview, ?) as suffix_score,
       rapidfuzz_partial_ratio      (preview, ?) as partial_score_prime,
       partial_score_prime * list_min([length(preview)/length(?), 1]) as partial_score,

       rapidfuzz_ratio              (primary_mime_type, ?) as type_score,

       ratio_score + partial_score + osa_score +
       lcs_score + prefix_score + suffix_score + type_score as score
    FROM events
		WHERE score IS NOT NULL AND %s
		ORDER BY score DESC
		LIMIT ?;
	`, whereClause)

	previews := slices.Repeat([]any{q.Fuzzy}, 7)

	args = append(args, q.Type, q.Limit)
	return query, append(previews, args...)
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
		if err != nil {
			return nil, err
		}
		db.mu.Lock()
		return res, nil
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

		// TODO: Do something with these value
		var PartialScorePrime, RatioScore, PartialScore, OsaScore,
			LcsScore, PrefixScore, SuffixScore, TypeScore float64

		err := rows.Scan(
			&i.ID,
			&i.MimeType,
			&i.Time,
			&i.Preview,
			&RatioScore,
			&OsaScore,
			&LcsScore,
			&PrefixScore,
			&SuffixScore,
			&TypeScore,
			&PartialScorePrime,
			&PartialScore,
			&i.Score,
		)
		if err != nil {
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
		err := rows.Scan(&i.ID, &i.MimeType, &i.Time, &i.Preview)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	err = rows.Close()
	if err != nil {
		return nil, err
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return items, nil
}
