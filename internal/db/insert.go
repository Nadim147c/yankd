package db

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/Nadim147c/yankd/internal/db/sqlc"
	"github.com/Nadim147c/yankd/internal/models"
)

// Insert inserts a clip to database.
func (db *DB) Insert(ctx context.Context, e models.Event) error {
	if len(e.Entries) == 0 {
		return errors.New("clipboard has no content")
	}
	slog.Debug("Inserting clip metadata into database")
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := db.queries.WithTx(tx)
	event, err := queries.CreateEvent(ctx, sqlc.CreateEventParams{
		PrimaryMimeType: e.PrimaryMimeType,
		Time:            e.Time,
	})
	if err != nil {
		return err
	}

	hashes := make(map[models.Hash]int64, len(e.Entries))

	for entry := range slices.Values(e.Entries) {
		var dataID int64

		// It's highly unlikely that there will be different content with same hash
		// on a same event.
		if i, ok := hashes[entry.Hash]; ok {
			dataID = i
			goto insertEntry
		}

		{
			dbDatas, err := queries.GetDatasByHash(ctx, entry.Hash)
			if err != nil {
				return fmt.Errorf("failed to find existing entries: %w", err)
			}

			slog.Debug("existing database datas", "count", len(dbDatas))

			// Collision aware insertiong by check the underlying data
			for dbEntry := range slices.Values(dbDatas) {
				textMatched := dbEntry.Text == entry.Text
				blobMatched := bytes.Equal(dbEntry.Blob, entry.Blob)

				// insert the entry when the underlying data already exists
				if (dbEntry.IsText && textMatched) || (!dbEntry.IsText && blobMatched) {
					dataID = dbEntry.ID
					slog.Debug("skipping insertiong",
						"data_id", dataID,
						"text_matched", textMatched,
						"blob_matched", blobMatched,
					)
					goto insertEntry
				}
			}
		}

		{
			data, err := queries.CreateData(ctx, sqlc.CreateDataParams{
				Hash:   entry.Hash,
				IsText: entry.IsText,
				Text:   entry.Text,
				Blob:   entry.Blob,
			})
			if err != nil {
				return fmt.Errorf("failed to insert data: %w", err)
			}
			slog.Debug("created a new data", "id", data.ID, "hash", data.Hash)
			dataID = data.ID
		}

	insertEntry:
		hashes[entry.Hash] = dataID

		slog.Debug("inserting entry", "event_id", event.ID, "mime_type", entry.MimeType, "data_id", dataID)
		entry, err := queries.CreateEntry(ctx, sqlc.CreateEntryParams{
			EventID:  event.ID,
			MimeType: entry.MimeType,
			DataID:   dataID,
		})
		if err != nil {
			return fmt.Errorf("failed to insert entry: %w", err)
		}
		slog.Debug("created a new entry",
			"event_id", entry.EventID,
			"mime_type", entry.MimeType,
			"data_id", entry.DataID,
		)
	}

	return tx.Commit()
}
