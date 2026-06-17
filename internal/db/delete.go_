package db

import "context"

// Delete deletes multiple items from database.
func (db *DB) Delete(ctx context.Context, ids []int64) (int64, error) {
	return db.queries.DeleteEvents(ctx, ids)
}

// Wipe deletes all items from database.
func (db *DB) Wipe(ctx context.Context) (int64, error) {
	return db.queries.DeleteAllEvents(ctx)
}
