package db

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Nadim147c/yankd/internal/db/binds"
	"github.com/Nadim147c/yankd/pkgs/clipboard"
	"github.com/glebarez/sqlite"
	"github.com/spf13/viper"
	"github.com/zeebo/xxh3"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

// GetDB returns creates a sqlite db in disk and return *gorm.DB.
func GetDB() (*gorm.DB, error) {
	var err error
	once.Do(func() {
		db, err = createDB()
	})
	return db, err
}

func createDB() (*gorm.DB, error) {
	dbDir := viper.GetString("database")
	if dbDir == "" {
		slog.Error("database directory is empty")
		return nil, errors.New("database directory can not be empty")
	}

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
		return clipboard.Clip{}, err
	}

	slog.Debug("successfully the clip", "id", clip.ID)
	return clipboard.Clip{}, nil
}

// Insert inserts given clip to database. Returns error on databse failure.
func Insert(ctx context.Context, clip *clipboard.Clip) error {
	slog.Debug(
		"inserting clip",
		"text-size", len(clip.Text),
		"blob-size", len(clip.Blob),
	)

	db, err := GetDB()
	if err != nil {
		slog.Error("failed to get database connection", "error", err)
		return err
	}

	if len(clip.Blob) != 0 {
		blobPath, err := CreateBlob(clip.Blob)
		if err != nil {
			slog.Error("failed to create blob", "error", err)
			return err
		}
		clip.BlobPath = blobPath
		clip.Blob = nil
	}

	var last clipboard.Clip
	if err := db.WithContext(ctx).Last(&last).Error; err != nil {
		slog.Error("failed to insert clip", "error", err)
		return err
	}

	if clip.Text == last.Text &&
		clip.Metadata == last.Metadata &&
		clip.URL == last.URL &&
		clip.BlobPath == last.BlobPath {
		slog.Debug("Ignoring duplicate clipboard item")
		return nil
	}

	if err := db.WithContext(ctx).Create(clip).Error; err != nil {
		slog.Error("failed to insert clip", "error", err)
		return err
	}

	slog.Debug("clip inserted successfully")
	return nil
}

// CreateBlob create a file containing the binary files in database/blob
// directory.
func CreateBlob(b []byte) (string, error) {
	sum := xxh3.Hash128(b).Bytes()
	id := hex.EncodeToString(sum[:])

	dbDir := viper.GetString("database")
	if dbDir == "" {
		slog.Error("database directory is empty")
		return "", errors.New("database directory can not be empty")
	}

	blobDir := filepath.Join(dbDir, "blob")
	if err := os.MkdirAll(blobDir, 0o755); err != nil {
		slog.Error(
			"failed to create blob directory",
			"path", blobDir,
			"error", err,
		)
		return "", fmt.Errorf("failed to create blob directory: %w", err)
	}

	path := filepath.Join(blobDir, id)
	if _, err := os.Stat(path); err == nil {
		slog.Debug("blob file already exists", "path", path)
		return path, nil
	}

	err := os.WriteFile(path, b, 0o644)
	if err != nil {
		slog.Error("failed to write blob file", "path", path, "error", err)
		return path, err
	}

	slog.Debug("blob file written", "path", path, "size", len(b))
	return path, nil
}
