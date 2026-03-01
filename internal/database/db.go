package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var (
	DB            *sql.DB
	isInitialized bool
)

// GetDefaultDBPath returns the OS-specific default database path
func GetDefaultDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".clippy", "clippy.db"), nil
}

// InitDB initializes the database with an optional custom path.
// If dbPath is empty, uses the default OS-specific path.
func InitDB(dbPath string) error {
	if DB != nil {
		if err := CloseDB(); err != nil {
			return fmt.Errorf("failed to close existing database connection: %w", err)
		}
	}

	// Resolve database path
	if dbPath == "" {
		var err error
		dbPath, err = GetDefaultDBPath()
		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open the database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database at %s: %w", dbPath, err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	isInitialized = true

	// Initialize schema silently
	if err := InitSchema(); err != nil {
		CloseDB()
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// CloseDB closes the database connection if it was opened
func CloseDB() error {
	if DB != nil {
		err := DB.Close()
		DB = nil
		isInitialized = false
		return err
	}
	return nil
}

// IsInitialized returns true if the database has been initialized
func IsInitialized() bool {
	return isInitialized && DB != nil
}
