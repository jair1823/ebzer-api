package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type ConnectConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

func GetConfig() ConnectConfig {
	config := ConnectConfig{
		User:     getEnvOrDefault("PGUSER", "postgres"),
		Password: getEnvOrDefault("PGPASSWORD", ""),
		Host:     getEnvOrDefault("PGHOST", "localhost"),
		Port:     getEnvOrDefault("PGPORT", "5432"),
		Database: getEnvOrDefault("PGDATABASE", "postgres"),
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
	if c.User == "" {
		return fmt.Errorf("PGUSER cannot be empty")
	}
	if c.Host == "" {
		return fmt.Errorf("PGHOST cannot be empty")
	}
	if c.Port == "" {
		return fmt.Errorf("PGPORT cannot be empty")
	}
	if c.Database == "" {
		return fmt.Errorf("PGDATABASE cannot be empty")
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
			log.Println("Successful connection to PostgreSQL")
			return db, nil
		}

		log.Printf("Attempt %d failed: %v", attempt, err)
		// Do not wait after the last attempt
		if attempt < maxRetries {
			log.Printf("Waiting %d seconds before the next attempt...", delaySeconds)
			time.Sleep(delaySeconds * time.Second)
		}
	}

	return nil, fmt.Errorf("Could not connect after %d attempts: %v", maxRetries, err)
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
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	log.Printf("Connecting to: postgres://%s:***@%s:%s/%s", config.User, config.Host, config.Port, config.Database)

	// Open connection with timeout
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("error opening connection: %w", err)
	}

	// Set connection configuration
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("error verifying connection (ping): %w", err)
	}

	return db, nil
}
