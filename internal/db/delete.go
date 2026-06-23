package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
)

// Delete deletes multiple items from database.
func (db *DB) Delete(ctx context.Context, ids ...uuid.UUID) (int64, error) {
	db.mu.Lock()
	defer db.SafeUnlock()

	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	const deleteEvents = `
    DELETE FROM events WHERE id IN [@UUID_ARRAY@];
  `
	query := strings.ReplaceAll(deleteEvents, "@UUID_ARRAY@", convertUUIDsToString(ids))

	res, err := tx.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	m, err := optimziseDatabse(ctx, tx)
	if err != nil {
		return 0, err
	}

	return n + m, tx.Commit()
}

func execBatch(ctx context.Context, tx *sql.Tx, queries ...string) (n int64, err error) {
	for _, query := range queries {
		res, err := tx.ExecContext(ctx, query)
		if err != nil {
			return n, err
		}
		m, err := res.RowsAffected()
		if err != nil {
			return n, err
		}
		n += m
	}
	return n, nil
}

func optimziseDatabse(ctx context.Context, tx *sql.Tx) (int64, error) {
	return execBatch(
		ctx, tx,
		`
    DELETE FROM events
    WHERE NOT EXISTS (
        SELECT 1 FROM entries
        WHERE entries.event_id = events.id
    );`,
		`
    DELETE FROM entries
    WHERE NOT EXISTS (
        SELECT 1 FROM events
        WHERE entries.event_id = events.id
    );`,
		`
    DELETE FROM contents
    WHERE NOT EXISTS (
        SELECT 1 FROM entries
        WHERE contents.id = entries.content_id
    );`,
	)
}

// Wipe deletes all items from database.
func (db *DB) Wipe(ctx context.Context) (int64, error) {
	db.mu.Lock()
	defer db.SafeUnlock()

	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	n, err := execBatch(
		ctx, tx,
		`DROP TABLE events;`,
		`DROP TABLE entries;`,
		`DROP TABLE contents;`,
		configureQuery,
	)
	if err != nil {
		return 0, err
	}
	return n, tx.Commit()
}
