package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"
	_ "github.com/mattn/go-sqlite3"
)

// DB is the global database instance
var (
	db   *sql.DB
	dbMu sync.RWMutex
)

// Config holds database configuration options
type Config struct {
	// Enabled determines if database persistence is active
	Enabled bool `json:"enabled"`
	// Path is the database file path (optional, uses default if empty)
	Path string `json:"path,omitempty"`
	// RetentionDays is how long to keep historical data (default 30)
	RetentionDays int `json:"retentionDays,omitempty"`
}

// DefaultConfig returns the default database configuration
func DefaultConfig() Config {
	return Config{
		Enabled:       true,
		Path:          "",
		RetentionDays: 30,
	}
}

// defaultDBPath returns the default database file path
func defaultDBPath() (string, error) {
	dataDir := filepath.Join(xdg.DataHome, "lethean-desktop")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create data directory: %w", err)
	}
	return filepath.Join(dataDir, "mining.db"), nil
}

// Initialize opens the database connection and creates tables
func Initialize(cfg Config) error {
	dbMu.Lock()
	defer dbMu.Unlock()

	if !cfg.Enabled {
		return nil
	}

	dbPath := cfg.Path
	if dbPath == "" {
		var err error
		dbPath, err = defaultDBPath()
		if err != nil {
			return err
		}
	}

	var err error
	db, err = sql.Open("sqlite3", dbPath+"?_journal=WAL&_timeout=5000")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite only supports one writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Create tables
	if err := createTables(); err != nil {
		db.Close()
		db = nil
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	dbMu.Lock()
	defer dbMu.Unlock()

	if db == nil {
		return nil
	}

	err := db.Close()
	db = nil
	return err
}

// IsInitialized returns true if the database is ready
func IsInitialized() bool {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return db != nil
}

// GetDB returns the database connection (for advanced queries).
//
// Deprecated: This function is unsafe for concurrent use because the returned
// pointer may become invalid if Close() is called by another goroutine after
// GetDB() returns. Use the dedicated query functions (InsertHashratePoint,
// GetHashrateHistory, etc.) instead, which handle locking internally.
func GetDB() *sql.DB {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return db
}

// createTables creates all required database tables
func createTables() error {
	schema := `
	-- Hashrate history table for storing miner performance data
	CREATE TABLE IF NOT EXISTS hashrate_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		miner_name TEXT NOT NULL,
		miner_type TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		hashrate INTEGER NOT NULL,
		resolution TEXT NOT NULL DEFAULT 'high',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Index for efficient queries by miner and time range
	CREATE INDEX IF NOT EXISTS idx_hashrate_miner_time
		ON hashrate_history(miner_name, timestamp DESC);

	-- Index for cleanup queries
	CREATE INDEX IF NOT EXISTS idx_hashrate_resolution_time
		ON hashrate_history(resolution, timestamp);

	-- Miner sessions table for tracking uptime
	CREATE TABLE IF NOT EXISTS miner_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		miner_name TEXT NOT NULL,
		miner_type TEXT NOT NULL,
		started_at DATETIME NOT NULL,
		stopped_at DATETIME,
		total_shares INTEGER DEFAULT 0,
		rejected_shares INTEGER DEFAULT 0,
		average_hashrate INTEGER DEFAULT 0
	);

	-- Index for session queries
	CREATE INDEX IF NOT EXISTS idx_sessions_miner
		ON miner_sessions(miner_name, started_at DESC);
	`

	_, err := db.Exec(schema)
	return err
}

// Cleanup removes old data based on retention settings
func Cleanup(retentionDays int) error {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	_, err := db.Exec(`
		DELETE FROM hashrate_history
		WHERE timestamp < ?
	`, cutoff)

	return err
}

// VacuumDB optimizes the database file size
func VacuumDB() error {
	dbMu.RLock()
	defer dbMu.RUnlock()

	if db == nil {
		return nil
	}

	_, err := db.Exec("VACUUM")
	return err
}
