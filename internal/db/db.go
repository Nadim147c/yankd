package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Nadim147c/yankd/internal/db/binds"
	"github.com/Nadim147c/yankd/pkg/clipboard"
	"github.com/cespare/xxhash/v2"
	"github.com/glebarez/sqlite"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	instance *gorm.DB
	once     sync.Once
)

// GetDB returns creates a sqlite db in disk and return *gorm.DB.
func GetDB() (*gorm.DB, error) {
	var err error
	once.Do(func() {
		instance, err = createDB()
	})
	return instance, err
}

func createDB() (*gorm.DB, error) {
	dbDir := viper.GetString("database")
	if dbDir == "" {
		slog.Error("database directory is empty")
		return nil, errors.New("database directory can not be empty")
	}

	slog.Info("databse initialization", "database-dir", dbDir)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		slog.Error(
			"failed to create database directory",
			"path", dbDir,
			"error", err,
		)
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	slog.Debug("database directory created", "path", dbDir)

	newLogger := logger.New(
		log.Default(),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  false,
		},
	)

	dbFilename := sqlite.Open(filepath.Join(dbDir, "history.db"))
	db, err := gorm.Open(dbFilename, &gorm.Config{Logger: newLogger})
	if err != nil {
		slog.Error("failed to open database", "error", err)
		return nil, err
	}

	// allow multiple process write without error
	if err := db.Raw(`PRAGMA busy_timeout = 5000;`).Error; err != nil {
		return db, err
	}
	if err := db.Raw(`PRAGMA journal_mode = WAL;`).Error; err != nil {
		return db, err
	}

	if err := db.AutoMigrate(&clipboard.Clip{}); err != nil {
		slog.Error("failed to auto migrate database", "error", err)
		return nil, err
	}

	slog.Info("database connected successfully")
	return db, nil
}

// Get returns clip for given id. Returns error if id not exists or db failure.
func Get(ctx context.Context, id uint) (clipboard.Clip, error) {
	slog.Debug("searching for clipboard", "id", id)

	db, err := GetDB()
	if err != nil {
		slog.Error("failed to get database connection", "error", err)
		return clipboard.Clip{}, err
	}

	clip, err := gorm.G[clipboard.Clip](db).
		Where(binds.Clip.ID.Eq(id)).
		First(ctx)
	if err != nil {
		slog.Error("failed to find clip", "id", id, "error", err)
		return clipboard.Clip{}, fmt.Errorf("failed to find clip: %v", err)
	}

	slog.Debug("successfully the clip", "id", clip.ID)
	return clip, nil
}

// Insert inserts given clip to database. Returns error on databse failure.
func Insert(ctx context.Context, clip clipboard.Clip) (clipboard.Clip, error) {
	slog.Debug(
		"inserting clip",
		"text-size", len(clip.Text),
		"blob-size", len(clip.Blob),
	)

	db, err := GetDB()
	if err != nil {
		slog.Error("failed to get database connection", "error", err)
		return clip, err
	}

	if len(clip.Blob) != 0 {
		blobHash, blobPath, err := CreateBlob(clip.Blob)
		if err != nil {
			slog.Error("failed to create blob", "error", err)
			return clip, err
		}
		clip.BlobPath = blobPath
		clip.BlobHash = blobHash
		clip.Blob = nil
	}

	clip.Hash = clipboard.HashClip(clip)

	dbClip, err := gorm.G[clipboard.Clip](db).
		Where(binds.Clip.Hash.Eq(clip.Hash)).
		First(ctx)
	if err == nil {
		slog.Debug("record already exists", "hash", clip.Hash)
		return dbClip, nil
	}

	if err := gorm.G[clipboard.Clip](db).Create(ctx, &clip); err != nil {
		slog.Error("failed to insert clip", "error", err)
		return clip, err
	}

	slog.Debug("clip inserted successfully")
	return clip, nil
}

// CreateBlob create a file containing the binary files in database/blob
// directory.
func CreateBlob(b []byte) (clipboard.Hash, string, error) {
	id := clipboard.Hash(xxhash.Sum64(b))

	dbDir := viper.GetString("database")
	if dbDir == "" {
		slog.Error("database directory is empty")
		return id, "", errors.New("database directory can not be empty")
	}

	blobDir := filepath.Join(dbDir, "blob")
	if err := os.MkdirAll(blobDir, 0o755); err != nil {
		slog.Error(
			"failed to create blob directory",
			"path", blobDir,
			"error", err,
		)
		return id, "", fmt.Errorf("failed to create blob directory: %w", err)
	}

	path := filepath.Join(blobDir, fmt.Sprint(id))
	if _, err := os.Stat(path); err == nil {
		slog.Debug("blob file already exists", "path", path)
		return id, path, nil
	}

	err := os.WriteFile(path, b, 0o644)
	if err != nil {
		slog.Error("failed to write blob file", "path", path, "error", err)
		return id, path, err
	}

	slog.Debug("blob file written", "path", path, "size", len(b))
	return id, path, nil
}
