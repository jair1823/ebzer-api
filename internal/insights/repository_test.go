package insights

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func newTestInsightsRepository(t *testing.T) Repository {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(`
		CREATE TABLE order_statuses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			is_final_status INTEGER DEFAULT 0
		);
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			description TEXT NOT NULL,
			amount_charged REAL NOT NULL,
			status_id INTEGER NOT NULL,
			entry_date TEXT NOT NULL,
			estimated_delivery_date TEXT,
			platform TEXT NOT NULL DEFAULT 'whatsapp',
			deleted_at TEXT
		);
		CREATE TABLE income (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			amount REAL NOT NULL,
			date TEXT NOT NULL,
			deleted_at TEXT
		);
		CREATE TABLE comercios (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
		CREATE TABLE expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			comercio_id INTEGER NOT NULL,
			amount REAL,
			date TEXT NOT NULL,
			deleted_at TEXT
		);
		CREATE TABLE expense_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			expense_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			unit_price REAL NOT NULL
		);

		INSERT INTO order_statuses (id, name, is_final_status)
		VALUES (1, 'new', 0), (2, 'completed', 1);
		INSERT INTO orders (id, description, amount_charged, status_id, entry_date, estimated_delivery_date, platform)
		VALUES
			(1, 'Active order', 100, 1, '2026-06-02', '2026-06-01', 'whatsapp'),
			(2, 'Paid order', 80, 2, '2026-06-03', '2026-06-15', 'instagram');
		INSERT INTO income (order_id, amount, date) VALUES (1, 25, '2026-06-04'), (2, 80, '2026-06-05');
		INSERT INTO comercios (id, name) VALUES (1, 'Universal'), (2, 'Mercado');
		INSERT INTO expenses (id, comercio_id, date) VALUES (1, 1, '2026-06-06');
		INSERT INTO expenses (id, comercio_id, amount, date) VALUES (2, 2, 45, '2026-06-07');
		INSERT INTO expense_items (expense_id, quantity, unit_price) VALUES (1, 2, 15);
	`); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return NewRepository(db)
}

func TestRepositoryGetSummary(t *testing.T) {
	repo := newTestInsightsRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	summary, err := repo.GetSummary(ctx, SummaryFilter{From: "2026-06-01", To: "2026-06-30"})
	if err != nil {
		t.Fatalf("GetSummary returned error: %v", err)
	}

	if summary.IncomeTotal != 105 {
		t.Fatalf("expected income 105, got %.2f", summary.IncomeTotal)
	}
	if summary.ExpenseTotal != 75 {
		t.Fatalf("expected expenses 75, got %.2f", summary.ExpenseTotal)
	}
	if summary.Profit != 30 {
		t.Fatalf("expected profit 30, got %.2f", summary.Profit)
	}
	if summary.PendingCollection != 75 {
		t.Fatalf("expected pending 75, got %.2f", summary.PendingCollection)
	}
	if summary.ActiveOrders != 1 {
		t.Fatalf("expected 1 active order, got %d", summary.ActiveOrders)
	}
	if summary.PaidCompletedOrders != 1 {
		t.Fatalf("expected 1 paid completed order, got %d", summary.PaidCompletedOrders)
	}
	if len(summary.SalesByPlatform) != 2 {
		t.Fatalf("expected sales for 2 platforms, got %#v", summary.SalesByPlatform)
	}
	if len(summary.TopExpenseMerchants) != 2 || summary.TopExpenseMerchants[0].Name != "Mercado" {
		t.Fatalf("expected Mercado top merchant, got %#v", summary.TopExpenseMerchants)
	}
}
