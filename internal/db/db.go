package db

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Nadim147c/yankd/pkgs/clipboard"
	"github.com/glebarez/sqlite"
	"github.com/spf13/viper"
	"github.com/zeebo/xxh3"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	dbDir := viper.GetString("database")
	if dbDir == "" {
		return nil, errors.New("database directory can not be empty")
	}

	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return nil, fmt.Errorf("Failed to create database directory")
	}

	dbFilename := sqlite.Open(filepath.Join(dbDir, "history.db"))

	db, err := gorm.Open(dbFilename)
	if err != nil {
		return nil, err
	}

	return db, db.AutoMigrate(&clipboard.Clip{})
}

func Insert(ctx context.Context, clip *clipboard.Clip) error {
	db, err := Connect()
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

	return gorm.G[clipboard.Clip](db).Create(ctx, clip)
}

func CreateBlob(b []byte) (string, error) {
	sum := xxh3.Hash128(b).Bytes()
	id := hex.EncodeToString(sum[:])
	dbDir := viper.GetString("database")
	if dbDir == "" {
		return "", errors.New("database directory can not be empty")
	}

	if err := os.MkdirAll(filepath.Join(dbDir, "blob"), 0o755); err != nil {
		return "", fmt.Errorf("Failed to create database directory")
	}

	path := filepath.Join(dbDir, "blob", id)
	if _, err := os.Stat(path); err == nil {
		return path, nil // file already exists
	}

	err := os.WriteFile(path, b, 0o644)
	return path, err
}
