package expenses

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestExpenseRepository(t *testing.T) (*sql.DB, Repository) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(`
		PRAGMA foreign_keys = ON;

		CREATE TABLE comercios (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL COLLATE NOCASE UNIQUE,
			description TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);

		CREATE TABLE products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			comercio_id INTEGER NOT NULL REFERENCES comercios(id) ON DELETE CASCADE,
			name TEXT NOT NULL COLLATE NOCASE,
			default_price REAL NOT NULL CHECK (default_price > 0),
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			UNIQUE (comercio_id, name)
		);

		CREATE TABLE expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			comercio_id INTEGER NOT NULL REFERENCES comercios(id) ON DELETE RESTRICT,
			date TEXT NOT NULL DEFAULT (date('now')),
			description TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);

		CREATE TABLE expense_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			expense_id INTEGER NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
			product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
			product_name TEXT NOT NULL,
			quantity REAL NOT NULL CHECK (quantity > 0),
			unit_price REAL NOT NULL CHECK (unit_price > 0),
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);

		INSERT INTO comercios (id, name, description) VALUES
			(1, 'Universal', 'Tienda principal'),
			(2, 'Mercado', NULL);
		INSERT INTO products (id, comercio_id, name, default_price) VALUES
			(1, 1, 'Tela', 1000),
			(2, 1, 'Hilo', 500),
			(3, 2, 'Bolsa', 100);
	`); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db, NewRepository(db)
}

func TestRepositoryCreateExpenseWithExistingAndNewProducts(t *testing.T) {
	db, repo := newTestExpenseRepository(t)
	ctx := context.Background()
	expenseDate := "2026-06-03"
	description := "Compra de materiales"
	productID := 1

	id, err := repo.Create(ctx, CreateExpenseDTO{
		ComercioID:  1,
		Date:        &expenseDate,
		Description: &description,
		Items: []CreateExpenseItemDTO{
			{ProductID: &productID, ProductName: "Tela", Quantity: 2.5, UnitPrice: 1200},
			{ProductName: "Cinta", Quantity: 3, UnitPrice: 250},
		},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	expense, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if expense == nil {
		t.Fatal("expected expense")
	}
	if expense.Total != 3750 {
		t.Fatalf("expected total 3750, got %v", expense.Total)
	}
	if len(expense.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(expense.Items))
	}
	if expense.Items[0].LineTotal != 3000 {
		t.Fatalf("expected first line total 3000, got %v", expense.Items[0].LineTotal)
	}

	var latestPrice float64
	if err := db.QueryRowContext(ctx, "SELECT default_price FROM products WHERE id = 1").Scan(&latestPrice); err != nil {
		t.Fatalf("select existing product price: %v", err)
	}
	if latestPrice != 1200 {
		t.Fatalf("expected product default price to update to 1200, got %v", latestPrice)
	}

	var createdProductCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE comercio_id = 1 AND name = 'Cinta'").Scan(&createdProductCount); err != nil {
		t.Fatalf("select created product count: %v", err)
	}
	if createdProductCount != 1 {
		t.Fatalf("expected new product to be created, got %d", createdProductCount)
	}
}

func TestRepositoryGetAllFiltersByDateAndComercio(t *testing.T) {
	_, repo := newTestExpenseRepository(t)
	ctx := context.Background()
	description := "Materiales"

	if _, err := repo.Create(ctx, CreateExpenseDTO{
		ComercioID:  1,
		Date:        stringPtr("2026-06-05"),
		Description: &description,
		Items:       []CreateExpenseItemDTO{{ProductName: "Tela", Quantity: 1, UnitPrice: 1000}},
	}); err != nil {
		t.Fatalf("create first expense: %v", err)
	}
	if _, err := repo.Create(ctx, CreateExpenseDTO{
		ComercioID: 2,
		Date:       stringPtr("2026-06-06"),
		Items:      []CreateExpenseItemDTO{{ProductName: "Bolsa", Quantity: 1, UnitPrice: 100}},
	}); err != nil {
		t.Fatalf("create second expense: %v", err)
	}

	from := "2026-06-01"
	to := "2026-06-30"
	comercioID := 1
	expenses, err := repo.GetAll(ctx, &from, &to, &comercioID)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(expenses) != 1 {
		t.Fatalf("expected 1 expense, got %d", len(expenses))
	}
	if expenses[0].Comercio == nil || expenses[0].Comercio.Name != "Universal" {
		t.Fatalf("expected comercio Universal, got %#v", expenses[0].Comercio)
	}
	if len(expenses[0].Items) != 1 {
		t.Fatalf("expected joined items, got %#v", expenses[0].Items)
	}
}

func TestRepositoryUpdatePreservesHistoricalItemPrice(t *testing.T) {
	db, repo := newTestExpenseRepository(t)
	ctx := context.Background()
	productID := 1

	id, err := repo.Create(ctx, CreateExpenseDTO{
		ComercioID: 1,
		Items: []CreateExpenseItemDTO{
			{ProductID: &productID, ProductName: "Tela", Quantity: 1, UnitPrice: 1000},
		},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE products SET default_price = 1500 WHERE id = 1"); err != nil {
		t.Fatalf("update product price: %v", err)
	}

	expense, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if expense.Items[0].UnitPrice != 1000 {
		t.Fatalf("expected historical unit price 1000, got %v", expense.Items[0].UnitPrice)
	}
}

func TestRepositoryDeleteBlocksUsedComerciosAndProducts(t *testing.T) {
	_, repo := newTestExpenseRepository(t)
	ctx := context.Background()
	productID := 1

	if _, err := repo.Create(ctx, CreateExpenseDTO{
		ComercioID: 1,
		Items: []CreateExpenseItemDTO{
			{ProductID: &productID, ProductName: "Tela", Quantity: 1, UnitPrice: 1000},
		},
	}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := repo.DeleteComercio(ctx, 1); err == nil || err.Error() != "comercio is in use" {
		t.Fatalf("expected comercio in use error, got %v", err)
	}
	if err := repo.DeleteProduct(ctx, 1); err == nil || err.Error() != "product is in use" {
		t.Fatalf("expected product in use error, got %v", err)
	}
	if err := repo.DeleteProduct(ctx, 2); err != nil {
		t.Fatalf("DeleteProduct unused returned error: %v", err)
	}
}

func stringPtr(value string) *string {
	return &value
}
