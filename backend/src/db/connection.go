package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB is the global database connection
var DB *sql.DB

// Config holds database configuration
type Config struct {
	DatabasePath string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		DatabasePath: getEnv("DATABASE_PATH", "./data/database.db"),
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  5 * time.Minute,
	}
}

// Initialize initializes the database connection and runs migrations
func Initialize(cfg *Config) error {
	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection with WAL mode and busy timeout for better concurrency
	db, err := sql.Open("sqlite", cfg.DatabasePath+"?_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool (set to 1 for SQLite to avoid lock contention)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable additional SQLite optimizations
	_, err = db.Exec("PRAGMA foreign_keys = ON; PRAGMA cache_size = -64000;")
	if err != nil {
		log.Printf("Warning: Failed to set SQLite pragmas: %v", err)
	}

	DB = db
	log.Printf("Database connection established: %s", cfg.DatabasePath)

	// Run migrations
	if err := runMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// runMigrations executes database migrations
func runMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	// Read migration file
	migrationSQL, err := os.ReadFile("database/migrations/001_init.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		log.Println("Closing database connection...")
		return DB.Close()
	}
	return nil
}

// Health checks database health
func Health() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return DB.Ping()
}

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
