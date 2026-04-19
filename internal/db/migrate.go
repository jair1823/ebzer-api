package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations executes all pending migrations
func RunMigrations(db *sql.DB, migrationsPath string) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("error getting applied migrations: %w", err)
	}

	// Get migration files
	migrations, err := getMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("error reading migration files: %w", err)
	}

	// Execute pending migrations
	for _, migration := range migrations {
		if applied[migration.Name] {
			log.Printf("⏭️  Skipping migration (already applied): %s", migration.Name)
			continue
		}

		log.Printf("🔄 Applying migration: %s", migration.Name)

		if err := executeMigration(db, migration); err != nil {
			return fmt.Errorf("error executing migration %s: %w", migration.Name, err)
		}

		if err := recordMigration(db, migration.Name); err != nil {
			return fmt.Errorf("error recording migration %s: %w", migration.Name, err)
		}

		log.Printf("✅ Successfully applied: %s", migration.Name)
	}

	log.Println("✅ All migrations completed successfully")
	return nil
}

type Migration struct {
	Name    string
	UpPath  string
	Content string
}

func createMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	return err
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func getMigrationFiles(migrationsPath string) ([]Migration, error) {
	var migrations []Migration

	err := filepath.Walk(migrationsPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".up.sql") {
			return nil
		}

		// Extract migration name (e.g., "000001_create_orders_table")
		filename := filepath.Base(path)
		name := strings.TrimSuffix(filename, ".up.sql")

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", path, err)
		}

		migrations = append(migrations, Migration{
			Name:    name,
			UpPath:  path,
			Content: string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by name (they're numbered)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Name < migrations[j].Name
	})

	return migrations, nil
}

func executeMigration(db *sql.DB, migration Migration) error {
	// SQLite doesn't support transactions with DDL well, but we try anyway
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(migration.Content); err != nil {
		return err
	}

	return tx.Commit()
}

func recordMigration(db *sql.DB, name string) error {
	_, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", name)
	return err
}
