package db

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/spf13/viper"
)

// DB is the database instance.
type DB struct {
	sql *sql.DB
	mu  sync.Mutex
}

var internalTestModeDoNotUse = false

// CreateDB creates a new database instance.
func CreateDB() (*DB, error) {
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, err
	}

	sqlDB, err := sql.Open("duckdb", dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		return nil, err
	}

	slog.Info("database connected successfully")
	db := new(DB)
	db.sql = sqlDB
	if err := db.initialize(); err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}
	return db, nil
}

func getDatabasePath() (string, error) {
	if internalTestModeDoNotUse {
		return ":memory:", nil
	}

	dbDir := viper.GetString("database")
	if dbDir == "" {
		slog.Error("database directory is empty")
		return "", errors.New("database directory can not be empty")
	}

	slog.Info("databse initialization", "database-dir", dbDir)
	if err := os.MkdirAll(dbDir, 0o750); err != nil {
		slog.Error(
			"failed to create database directory",
			"path", dbDir,
			"error", err,
		)
		return "", fmt.Errorf("failed to create database directory: %w", err)
	}
	slog.Debug("database directory created", "path", dbDir)
	return filepath.Join(dbDir, "history-duckdb.db"), nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.sql.Close()
}

//go:embed configure.sql
var configureQuery string

// initialize initializes the database connection and full-text search.
func (db *DB) initialize() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, err := db.sql.Exec(configureQuery); err != nil { //nolint:noctx
		return err
	}
	return nil
}
