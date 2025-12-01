package db

import (
	"context"
	"fmt"

	"github.com/Nadim147c/yankd/pkgs/clipboard"
	"gorm.io/gorm"
)

// InitializeFTS sets up the FTS5 virtual table and triggers
func InitializeFTS(db *gorm.DB) error {
	// 1️⃣ Create FTS5 virtual table for clips
	if err := db.Exec(`
        CREATE VIRTUAL TABLE IF NOT EXISTS clip_index USING FTS5(
			text,
			url,
			metadata,
			content='clips',
			content_rowid='id'
        );
    `).Error; err != nil {
		return fmt.Errorf("failed to create FTS5 table: %w", err)
	}

	// 2️⃣ Create triggers to keep index in sync
	triggers := []string{
		// Insert
		`CREATE TRIGGER IF NOT EXISTS clip_ai AFTER INSERT ON clips BEGIN
            INSERT INTO clip_index(rowid, text, url, metadata)
            VALUES (new.id, new.text, new.url, new.metadata);
        END;`,
		// Update
		`CREATE TRIGGER IF NOT EXISTS clip_au AFTER UPDATE ON clips BEGIN
            UPDATE clip_index SET text=new.text, text=new.url, metadata=new.metadata
            WHERE rowid=new.id;
        END;`,
		// Delete
		`CREATE TRIGGER IF NOT EXISTS clip_ad AFTER DELETE ON clips BEGIN
            DELETE FROM clip_index WHERE rowid=old.id;
        END;`,
	}

	for _, t := range triggers {
		if err := db.Exec(t).Error; err != nil {
			return fmt.Errorf("failed to create trigger: %w", err)
		}
	}

	return nil
}

// RebuildIndex rebuilds the FTS index for all existing rows
func RebuildIndex(db *gorm.DB) error {
	return db.Exec(`INSERT INTO clip_index(clip_index) VALUES('rebuild')`).Error
}

// FlexibleSearch searches text + metadata with FTS5 and fallback LIKE
func FlexibleSearch(ctx context.Context, db *gorm.DB, query string) ([]clipboard.Clip, error) {
	var clips []clipboard.Clip

	ftsQuery := fmt.Sprintf("%s* OR metadata:%s*", query, query)

	// 1️⃣ Try FTS5 search first
	err := db.WithContext(ctx).
		Raw(`SELECT clips.* FROM clips 
             JOIN clip_index ON clip_index.rowid = clips.id 
             WHERE clip_index MATCH ?`, ftsQuery).
		Scan(&clips).Error

	if err == nil && len(clips) > 0 {
		return clips, nil
	}

	// 2️⃣ Fallback to normal LIKE search
	return gorm.G[clipboard.Clip](db).
		Where("text LIKE ? OR metadata LIKE ?", "%"+query+"%", "%"+query+"%").
		Find(ctx)
}
