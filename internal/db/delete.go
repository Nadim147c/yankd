package db

import (
	"context"
	"errors"
	"os"
	"slices"

	"github.com/Nadim147c/yankd/internal/db/binds"
	"github.com/Nadim147c/yankd/pkg/clipboard"
	"gorm.io/gorm"
)

// Delete deletes a multiple from database.
func Delete(ctx context.Context, id []uint) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}

	clips, err := gorm.G[clipboard.Clip](db).
		Where(binds.Clip.ID.In(id...)).
		Find(ctx)
	if err != nil {
		return 0, err
	}

	n, err := gorm.G[clipboard.Clip](db).
		Where(binds.Clip.ID.In(id...)).
		Delete(ctx)
	if err != nil {
		return n, err
	}

	if err := rebuildIndex(db); err != nil {
		return n, err
	}

	var blobErrs []error
	for clip := range slices.Values(clips) {
		if clip.BlobPath != "" {
			blobErrs = append(blobErrs, os.Remove(clip.BlobPath))
		}
	}

	return n, errors.Join(blobErrs...)
}
