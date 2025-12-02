package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nadim147c/yankd/internal/db/binds"
	"github.com/Nadim147c/yankd/pkgs/clipboard"
	"gorm.io/gorm"
)

// InitializeFTS sets up the FTS5 virtual table and triggers
func InitializeFTS() error {
	slog.Debug("Initializing FTS5")

	db, err := GetDB()
	if err != nil {
		return err
	}

	// Create FTS5 virtual table for clips
	if err := db.Exec(`
    CREATE VIRTUAL TABLE IF NOT EXISTS clip_index USING FTS5(
			text,
			url,
			metadata,
			content='clips',
			content_rowid='id'
    );
    `).Error; err != nil {
		slog.Error("failed to create FTS5 table", "error", err)
		return fmt.Errorf("failed to create FTS5 table: %w", err)
	}
	slog.Debug("FTS5 table created")

	// Create triggers to keep index in sync
	triggers := []string{
		// Insert
		`CREATE TRIGGER IF NOT EXISTS clip_ai AFTER INSERT ON clips BEGIN
            INSERT INTO clip_index(rowid, text, url, metadata)
            VALUES (new.id, new.text, new.url, new.metadata);
        END;`,
		// Update
		`CREATE TRIGGER IF NOT EXISTS clip_au AFTER UPDATE ON clips BEGIN
            UPDATE clip_index SET text=new.text, url=new.url, metadata=new.metadata
            WHERE rowid=new.id;
        END;`,
		// Delete
		`CREATE TRIGGER IF NOT EXISTS clip_ad AFTER DELETE ON clips BEGIN
            DELETE FROM clip_index WHERE rowid=old.id;
        END;`,
	}

	for i, t := range triggers {
		if err := db.Exec(t).Error; err != nil {
			slog.Error("failed to create trigger", "trigger", i, "error", err)
			return fmt.Errorf("failed to create trigger: %w", err)
		}
	}
	slog.Debug("FTS5 triggers created", "count", len(triggers))

	return nil
}

// rebuildIndex rebuilds the FTS index for all existing rows
func rebuildIndex(db *gorm.DB) error {
	slog.Debug("rebuilding FTS index")

	err := db.Exec("INSERT INTO clip_index(clip_index) VALUES('rebuild')").Error
	if err != nil {
		slog.Error("failed to rebuild FTS index", "error", err)
		return err
	}

	slog.Debug("FTS index rebuilt successfully")
	return nil
}

// Search searches runs full-text serach in database and returns matched items.
func Search(
	ctx context.Context,
	query string,
	limit int,
	sync bool,
) ([]clipboard.Clip, error) {
	slog.Debug("searching clips", "query", query)

	db, err := GetDB()
	if err != nil {
		slog.Error("failed to get database connection", "error", err)
		return nil, err
	}

	if sync {
		if err := rebuildIndex(db); err != nil {
			slog.Error("failed to rebuild FTS index", "error", err)
			return nil, err
		}
	}

	slog.Debug("starting flexible search", "query", query)

	var clips []clipboard.Clip
	ftsQuery := fmt.Sprintf(
		"%s* OR metadata:%s* OR url:%s*",
		query, query, query,
	)

	// Try FTS5 search first
	err = db.WithContext(ctx).Raw(`SELECT clips.* FROM clips
    JOIN clip_index ON clip_index.rowid = clips.id
    WHERE clip_index MATCH ?
    ORDER BY rank
    LIMIT ?
    `, ftsQuery, limit).
		Scan(&clips).Error

	if err == nil && len(clips) > 0 {
		slog.Debug(
			"FTS5 search succeeded",
			"query", query,
			"results", len(clips),
		)
		return clips, nil
	}

	if err != nil {
		slog.Debug(
			"FTS5 search failed, falling back to LIKE",
			"query", query,
			"error", err,
		)
	} else {
		slog.Debug(
			"FTS5 search returned no results, falling back to LIKE",
			"query", query,
		)
	}

	// Fallback to normal LIKE search
	likeQuery := "%" + query + "%"
	if err := db.WithContext(ctx).
		Where(binds.Clip.Text.Like(likeQuery)).
		Or(binds.Clip.Metadata.Like(likeQuery)).
		Or(binds.Clip.URL.Like(likeQuery)).
		Limit(limit).
		Find(&clips).Error; err != nil {
		slog.Error("fallback LIKE search failed", "query", query, "error", err)
		return nil, err
	}

	slog.Debug(
		"fallback LIKE search succeeded",
		"query", query,
		"results", len(clips),
	)
	return clips, nil
}
