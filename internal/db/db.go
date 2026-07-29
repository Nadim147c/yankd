package db

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
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
		slog.Error("failed to get database path", "error", err)
		os.Exit(1)
	}

	sqlDB, err := sql.Open("duckdb", dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	slog.Info("database connected successfully")
	db := new(DB)
	db.sql = sqlDB
	if err := db.initialize(); err != nil {
		slog.Error("database initialization failed", "error", err)
		os.Exit(1)
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

	slog.Info("database initialization", "database-dir", dbDir)
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

// reconnect will close existing database connection and
// reconnect to that database.
func (db *DB) reconnect() {
	db.mu.Lock()
	defer db.SafeUnlock()

	err := db.sql.Close()
	if err != nil {
		slog.Error("failed to close existing", "error", err)
		os.Exit(1)
	}
	dbPath, err := getDatabasePath()
	if err != nil {
		slog.Error("failed to get database path", "error", err)
		os.Exit(1)
	}

	sqlDB, err := sql.Open("duckdb", dbPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	slog.Info("database connected successfully")
	db.sql = sqlDB
	if err := db.initialize(); err != nil {
		slog.Error("database initialization failed", "error", err)
		os.Exit(1)
	}
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.sql.Close()
}

// SafeUnlock tries to unlock underlying mutex.
// Return true if mutex was locked.
func (db *DB) SafeUnlock() (unlocked bool) {
	locked := db.mu.TryLock()
	db.mu.Unlock()
	return !locked
}

//go:embed configure.sql
var configureQuery string

// initialize initializes the database connection and full-text search.
func (db *DB) initialize() error {
	_, err := db.sql.Exec(configureQuery) //nolint:noctx
	return err
}
