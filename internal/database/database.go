package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Database wraps the sql.DB connection
type Database struct {
	DB *sql.DB
}

// Config holds database connection parameters
type Config struct {
	DataDir string
}

// New creates a new SQLite database connection
func New(cfg Config) (*Database, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(cfg.DataDir, "cupax.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{DB: db}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.DB.Close()
}

// InitSchema initializes the database schema
func (d *Database) InitSchema() error {
	schema := `
	-- Create analyses table
	CREATE TABLE IF NOT EXISTS analyses (
		id TEXT PRIMARY KEY,
		filename TEXT NOT NULL,
		file_hash_sha256 TEXT NOT NULL UNIQUE,
		status TEXT NOT NULL DEFAULT 'running',
		submitted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME,
		report_json TEXT,
		error_message TEXT,
		CHECK (status IN ('running', 'completed', 'error'))
	);

	-- Create index on status for faster filtering
	CREATE INDEX IF NOT EXISTS idx_analyses_status ON analyses(status);

	-- Create index on submitted_at for sorting
	CREATE INDEX IF NOT EXISTS idx_analyses_submitted_at ON analyses(submitted_at DESC);

	-- Create index on file_hash_sha256 for deduplication
	CREATE INDEX IF NOT EXISTS idx_analyses_file_hash ON analyses(file_hash_sha256);

	-- Create whitelist table
	CREATE TABLE IF NOT EXISTS whitelists (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		value TEXT NOT NULL,
		description TEXT,
		is_regex INTEGER NOT NULL DEFAULT 0,
		enabled INTEGER NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CHECK (type IN ('process', 'domain', 'ip', 'registry'))
	);

	-- Create index on type for faster filtering
	CREATE INDEX IF NOT EXISTS idx_whitelists_type ON whitelists(type);

	-- Create index on enabled for faster filtering
	CREATE INDEX IF NOT EXISTS idx_whitelists_enabled ON whitelists(enabled);

	-- Create unique index on type+value to prevent duplicates
	CREATE UNIQUE INDEX IF NOT EXISTS idx_whitelists_type_value ON whitelists(type, value);
	`

	_, err := d.DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}
