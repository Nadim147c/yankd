package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Nadim147c/yankd/pkg/clipboard"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Wipe delete all enteries from databse without deleting database and tables.
func Wipe(ctx context.Context) (int, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}

	n, err := gorm.G[clipboard.Clip](db).Where("true").Delete(ctx)
	if err != nil {
		return n, err
	}

	if err := rebuildIndex(db); err != nil {
		return n, err
	}

	dbDir := viper.GetString("database")
	if dbDir == "" {
		slog.Error("database directory is empty")
		return n, errors.New("database directory can not be empty")
	}

	blobDir := filepath.Join(dbDir, "blob")
	if err := os.RemoveAll(blobDir); err != nil {
		slog.Error(
			"failed to delete blob directory",
			"path", blobDir,
			"error", err,
		)
		return n, fmt.Errorf("failed to create blob directory: %w", err)
	}

	return n, nil
}
