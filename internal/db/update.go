package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UpdateEventTime updates the timestamp of a clipboard event.
func (db *DB) UpdateEventTime(ctx context.Context, id uuid.UUID, t time.Time) error {
	db.mu.Lock()
	defer db.SafeUnlock()
	_, err := db.sql.ExecContext(ctx, `UPDATE events SET time = ? WHERE id = ?`, t, id)
	return err
}
