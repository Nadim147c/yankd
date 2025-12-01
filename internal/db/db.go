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

func getDB() (*gorm.DB, error) {
	var err error
	once.Do(func() {
		db, err = Connect()
	})
	return db, err
}

func Connect() (*gorm.DB, error) {
	dbDir := viper.GetString("database")
	if dbDir == "" {
		slog.Error("database directory is empty")
		return nil, errors.New("database directory can not be empty")
	}

	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		slog.Error("failed to create database directory", "path", dbDir, "error", err)
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

	if err := db.AutoMigrate(&clipboard.Clip{}); err != nil {
		slog.Error("failed to auto migrate database", "error", err)
		return nil, err
	}
	slog.Info("database connected successfully")

	return db, nil
}

func Search(ctx context.Context, query string) ([]clipboard.Clip, error) {
	slog.Debug("searching clips", "query", query)

	db, err := getDB()
	if err != nil {
		slog.Error("failed to get database connection", "error", err)
		return nil, err
	}

	if err := InitializeFTS(db); err != nil {
		slog.Error("failed to initialize FTS", "error", err)
		return nil, err
	}

	if err := RebuildIndex(db); err != nil {
		slog.Error("failed to rebuild FTS index", "error", err)
		return nil, err
	}

	results, err := FlexibleSearch(context.Background(), db, query)
	if err != nil {
		slog.Error("search failed", "query", query, "error", err)
		return nil, err
	}

	slog.Debug("search completed", "query", query, "results", len(results))
	return results, nil
}

func Get(ctx context.Context, id uint) (*clipboard.Clip, error) {
	slog.Debug("searching for clipboard", "id", id)

	db, err := getDB()
	if err != nil {
		slog.Error("failed to get database connection", "error", err)
		return nil, err
	}

	clip, err := gorm.G[clipboard.Clip](db).Where(binds.Clip.ID.Eq(id)).First(ctx)
	if err != nil {
		slog.Error("failed to find clip", "id", id, "error", err)
		return nil, err
	}

	slog.Debug("successfully the clip", "id", clip.ID)
	return &clip, nil
}

func Insert(ctx context.Context, clip *clipboard.Clip) error {
	slog.Debug("inserting clip", "text-size", len(clip.Text), "blob-size", len(clip.Blob))

	db, err := getDB()
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

	if err := db.WithContext(ctx).Create(clip).Error; err != nil {
		slog.Error("failed to insert clip", "error", err)
		return err
	}

	slog.Debug("clip inserted successfully")
	return nil
}

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
		slog.Error("failed to create blob directory", "path", blobDir, "error", err)
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
