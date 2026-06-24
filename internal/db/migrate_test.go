package db

import (
	"database/sql"
	"os"
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

func TestExpenseItemQuantityIntegerMigrationRoundsUpDecimals(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(`
		CREATE TABLE expenses (id INTEGER PRIMARY KEY);
		CREATE TABLE products (id INTEGER PRIMARY KEY);
		CREATE TABLE expense_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			expense_id INTEGER NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
			product_name TEXT NOT NULL,
			quantity REAL NOT NULL CHECK (quantity > 0),
			unit_price REAL NOT NULL CHECK (unit_price > 0),
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);
		CREATE INDEX idx_expense_items_expense_id ON expense_items (expense_id);
		CREATE INDEX idx_expense_items_product_id ON expense_items (product_id);

		INSERT INTO expenses (id) VALUES (1);
		INSERT INTO products (id) VALUES (1);
		INSERT INTO expense_items (id, expense_id, product_id, product_name, quantity, unit_price, created_at)
		VALUES
			(1, 1, 1, 'Tela', 1.9, 1200, '2026-06-01'),
			(2, 1, 1, 'Hilo', 2.0, 500, '2026-06-01');
	`); err != nil {
		t.Fatalf("create old schema: %v", err)
	}

	content, err := os.ReadFile("migrations/000016_make_expense_item_quantity_integer.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := db.Exec(string(content)); err != nil {
		t.Fatalf("execute migration: %v", err)
	}

	var roundedQuantity int
	if err := db.QueryRow("SELECT quantity FROM expense_items WHERE id = 1").Scan(&roundedQuantity); err != nil {
		t.Fatalf("select rounded quantity: %v", err)
	}
	if roundedQuantity != 2 {
		t.Fatalf("expected 1.9 to round up to 2, got %d", roundedQuantity)
	}

	var wholeQuantity int
	if err := db.QueryRow("SELECT quantity FROM expense_items WHERE id = 2").Scan(&wholeQuantity); err != nil {
		t.Fatalf("select whole quantity: %v", err)
	}
	if wholeQuantity != 2 {
		t.Fatalf("expected 2.0 to remain 2, got %d", wholeQuantity)
	}

	var quantityType string
	if err := db.QueryRow("SELECT type FROM pragma_table_info('expense_items') WHERE name = 'quantity'").Scan(&quantityType); err != nil {
		t.Fatalf("select quantity type: %v", err)
	}
	if quantityType != "INTEGER" {
		t.Fatalf("expected quantity column to be INTEGER, got %s", quantityType)
	}
}
