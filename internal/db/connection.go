package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ConnectConfig struct {
	DatabasePath string
}

func GetConfig() ConnectConfig {
	config := ConnectConfig{
		DatabasePath: getEnvOrDefault("SQLITE_DB_PATH", "./data/ebzer.db"),
	}
	return config
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c ConnectConfig) Validate() error {
	if c.DatabasePath == "" {
		return fmt.Errorf("SQLITE_DB_PATH cannot be empty")
	}
	return nil
}

func ConnectWithRetry(maxRetries int, delaySeconds time.Duration) (*sql.DB, error) {
	config := GetConfig()

	// Validar configuración
	if err := config.Validate(); err != nil {
		log.Printf("❌ Invalid configuration: %v", err)
		return nil, err
	}

	var db *sql.DB
	var err error

	// try to connect with retries
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Connection attempt %d/%d...", attempt, maxRetries)

		db, err = connect(config)
		if err == nil {
			log.Println("✅ Successful connection to SQLite")
			return db, nil
		}

		log.Printf("Attempt %d failed: %v", attempt, err)
		// Do not wait after the last attempt
		if attempt < maxRetries {
			log.Printf("Waiting %d seconds before the next attempt...", delaySeconds)
			time.Sleep(delaySeconds * time.Second)
		}
	}

	return nil, fmt.Errorf("could not connect after %d attempts: %v", maxRetries, err)
}

// Connect tries to connect once
func Connect() (*sql.DB, error) {
	config := GetConfig()

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Printf("❌ Invalid configuration: %v", err)
		return nil, err
	}

	return connect(config)
}

// connect is the internal function that performs the connection
func connect(config ConnectConfig) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(config.DatabasePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error creating database directory: %w", err)
	}

	log.Printf("Connecting to SQLite: %s", config.DatabasePath)

	// Open SQLite connection
	// _loc=auto enables automatic timestamp parsing from TEXT columns
	db, err := sql.Open("sqlite3", config.DatabasePath+"?_foreign_keys=on&_journal_mode=WAL&_loc=auto")
	if err != nil {
		return nil, fmt.Errorf("error opening connection: %w", err)
	}

	// SQLite specific configuration
	db.SetMaxOpenConns(1) // SQLite works best with single connection
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("error verifying connection (ping): %w", err)
	}

	// Enable foreign keys (extra safety)
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	return db, nil
}
