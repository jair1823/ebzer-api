package db

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestRunMigrationsOnFreshDatabase(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := RunMigrations(db, "migrations"); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	for _, table := range []string{"orders", "income", "expenses", "agenda_items"} {
		var count int
		query := "SELECT COUNT(*) FROM pragma_table_info($1) WHERE name = 'deleted_at'"
		if err := db.QueryRow(query, table).Scan(&count); err != nil {
			t.Fatalf("query %s deleted_at: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("expected %s.deleted_at to exist", table)
		}
	}
}
