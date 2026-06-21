package db

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Nadim147c/yankd/internal/models"
	"github.com/google/uuid"
)

type content struct {
	ID     uuid.UUID
	Hash   models.Hash
	Blob   []byte
	IsText bool
}

func getContentsByHash(ctx context.Context, tx *sql.Tx, h models.Hash) ([]content, error) {
	slog.Debug("db.getContentsByHash", "hash", h)
	const getContentsByHash = `
    SELECT id, hash, is_text, blob FROM contents
    WHERE hash = ?
  `
	rows, err := tx.QueryContext(ctx, getContentsByHash, h)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []content{}
	for rows.Next() {
		var i content
		if err := rows.Scan(&i.ID, &i.Hash, &i.IsText, &i.Blob); err != nil {
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

// Insert inserts a clip to database.
func (db *DB) Insert(ctx context.Context, e *models.ClipboardEvent) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	hasUpdatedIndex.Store(false)

	if len(e.Entries) == 0 {
		return errors.New("clipboard has no content")
	}
	slog.Debug("Inserting clip metadata into database")
	tx, err := db.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const createEvent = `
    -- name: CreateEvent :one
    INSERT INTO events (primary_mime_type, time, preview)
    VALUES (?, ?, ?)
    RETURNING id
  `
	err = tx.
		QueryRowContext(ctx, createEvent, e.MimeType, e.Time, e.Preview).
		Scan(&e.ID)
	if err != nil {
		return err
	}

	hashes := make(map[models.Hash]uuid.UUID, len(e.Entries))

	for _, entry := range e.Entries {
		var contentID uuid.UUID

		// It's highly unlikely that there will be different content with same hash
		// on a same event.
		if i, ok := hashes[entry.Hash]; ok {
			contentID = i
			goto insertEntry
		}

		{
			dbContents, err := getContentsByHash(ctx, tx, entry.Hash)
			if err != nil {
				return fmt.Errorf("failed to find existing entries: %w", err)
			}

			slog.Debug("existing database contents", "count", len(dbContents))

			// Collision aware insertiong by check the underlying content
			for _, dbEntry := range dbContents {
				textMatch := dbEntry.IsText == entry.IsText
				blobMatched := bytes.Equal(dbEntry.Blob, entry.Blob)
				// insert the entry when the underlying content already exists
				if textMatch && blobMatched {
					contentID = dbEntry.ID
					slog.Debug(
						"skipping insertiong",
						"content_id", contentID,
						"text_matched", dbEntry.IsText == entry.IsText,
						"blob_matched", blobMatched,
					)
					goto insertEntry
				}
			}
		}

		{
			const createContent = `
        -- name: CreateContent :one
        INSERT INTO contents (hash, is_text, blob)
        VALUES (?, ?, ?)
        RETURNING id
      `
			err := db.sql.
				QueryRowContext(ctx, createContent, entry.Hash, entry.IsText, entry.Blob).
				Scan(&contentID)
			if err != nil {
				return fmt.Errorf("failed to insert content: %w", err)
			}
			slog.Debug("created a new content", "id", contentID, "hash", entry.Hash)
		}

	insertEntry:
		hashes[entry.Hash] = contentID

		slog.Debug("inserting entry", "event_id", e.ID, "mime_type", entry.MimeType, "content_id", contentID)
		const createEntry = `
      INSERT INTO entries (event_id, mime_type, content_id)
      VALUES (?, ?, ?)
    `
		_, err = db.sql.ExecContext(ctx, createEntry, e.ID, entry.MimeType, contentID)
		if err != nil {
			return fmt.Errorf("failed to insert entry: %w", err)
		}
	}

	return tx.Commit()
}
