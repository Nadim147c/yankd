package db

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/Nadim147c/yankd/pkgs/clipboard"
	"github.com/glebarez/sqlite"
	"github.com/spf13/viper"
	"github.com/zeebo/xxh3"
	"gorm.io/gorm"
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
		return nil, errors.New("database directory can not be empty")
	}

	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	dbFilename := sqlite.Open(filepath.Join(dbDir, "history.db"))
	db, err := gorm.Open(dbFilename)
	if err != nil {
		return nil, err
	}

	return db, db.AutoMigrate(&clipboard.Clip{})
}

func Search(ctx context.Context, query string) ([]clipboard.Clip, error) {
	db, err := getDB()
	if err != nil {
		return nil, err
	}

	// Step 1: Create FTS table + triggers
	if err := InitializeFTS(db); err != nil {
		return nil, err
	}

	// Step 2: Rebuild index for existing clips (optional)
	if err := RebuildIndex(db); err != nil {
		return nil, err
	}

	// Step 3: Perform search
	return FlexibleSearch(context.Background(), db, query)
}

func Insert(ctx context.Context, clip *clipboard.Clip) error {
	db, err := getDB()
	if err != nil {
		return err
	}

	if len(clip.Blob) != 0 {
		blobPath, err := CreateBlob(clip.Blob)
		if err != nil {
			return err
		}
		slog.Info("Create binary blob file", "path", blobPath)
		clip.BlobPath = blobPath
		clip.Blob = nil
	}

	return db.WithContext(ctx).Create(clip).Error
}

func CreateBlob(b []byte) (string, error) {
	sum := xxh3.Hash128(b).Bytes()
	id := hex.EncodeToString(sum[:])

	dbDir := viper.GetString("database")
	if dbDir == "" {
		return "", errors.New("database directory can not be empty")
	}

	if err := os.MkdirAll(filepath.Join(dbDir, "blob"), 0o755); err != nil {
		return "", fmt.Errorf("failed to create blob directory: %w", err)
	}

	path := filepath.Join(dbDir, "blob", id)
	if _, err := os.Stat(path); err == nil {
		return path, nil // file already exists
	}

	err := os.WriteFile(path, b, 0o644)
	return path, err
}
