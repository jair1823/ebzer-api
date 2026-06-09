package orders

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func newTestOrderRepository(t *testing.T) (*sql.DB, Repository) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	schema := `
	CREATE TABLE order_statuses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		display_name TEXT NOT NULL,
		color TEXT DEFAULT '#6B7280',
		order_position INTEGER NOT NULL,
		is_system_status INTEGER DEFAULT 0,
		is_final_status INTEGER DEFAULT 0,
		is_active INTEGER DEFAULT 1,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		description TEXT NOT NULL,
		amount_charged REAL NOT NULL,
		status_id INTEGER NOT NULL DEFAULT 1 REFERENCES order_statuses(id),
		entry_date TEXT NOT NULL DEFAULT (datetime('now')),
		estimated_delivery_date TEXT,
		delivery_type TEXT NOT NULL DEFAULT 'pickup',
		client_name TEXT,
		client_phone TEXT,
		notes TEXT,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	INSERT INTO order_statuses (id, name, display_name, color, order_position, is_system_status, is_final_status)
	VALUES
		(1, 'new', 'New', '#3B82F6', 1, 1, 0),
		(2, 'completed', 'Completed', '#10B981', 100, 1, 1);

	INSERT INTO orders (description, amount_charged, status_id, entry_date, delivery_type, client_name)
	VALUES
		('First order', 25.50, 1, '2026-01-01 10:00:00', 'pickup', 'Ana'),
		('Second order', 80.00, 2, '2026-01-02 10:00:00', 'delivery', 'Luis');
	`

	if _, err := db.Exec(schema); err != nil {
		if strings.Contains(err.Error(), "go-sqlite3 requires cgo") {
			t.Skipf("skipping SQLite repository test: %v", err)
		}
		t.Fatalf("create test schema: %v", err)
	}

	return db, NewRepository(db)
}

func TestRepositoryGetAllPopulatesStatusWithSingleConnection(t *testing.T) {
	_, repo := newTestOrderRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	orders, err := repo.GetAll(ctx, nil, nil, nil)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}

	byDescription := map[string]Order{}
	for _, order := range orders {
		byDescription[order.Description] = order
	}

	first := byDescription["First order"]
	if first.Status == nil {
		t.Fatal("expected First order to include status")
	}
	if first.Status.Name != "new" {
		t.Fatalf("expected First order status new, got %q", first.Status.Name)
	}

	second := byDescription["Second order"]
	if second.Status == nil {
		t.Fatal("expected Second order to include status")
	}
	if second.Status.Name != "completed" {
		t.Fatalf("expected Second order status completed, got %q", second.Status.Name)
	}
}

func TestRepositoryGetAllFiltersByStatusIDAndReturnsEmptyList(t *testing.T) {
	_, repo := newTestOrderRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	completedStatusID := 2
	orders, err := repo.GetAll(ctx, &completedStatusID, nil, nil)
	if err != nil {
		t.Fatalf("GetAll filtered returned error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 completed order, got %d", len(orders))
	}
	if orders[0].Status == nil || orders[0].Status.Name != "completed" {
		t.Fatalf("expected completed status, got %#v", orders[0].Status)
	}

	missingStatusID := 999
	orders, err = repo.GetAll(ctx, &missingStatusID, nil, nil)
	if err != nil {
		t.Fatalf("GetAll empty filter returned error: %v", err)
	}
	if len(orders) != 0 {
		t.Fatalf("expected empty order list, got %d", len(orders))
	}
}

func TestRepositoryGetByIDPopulatesStatusWithJoin(t *testing.T) {
	_, repo := newTestOrderRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	order, err := repo.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if order == nil {
		t.Fatal("expected order")
	}
	if order.Status == nil {
		t.Fatal("expected order status")
	}
	if order.Status.Name != "new" {
		t.Fatalf("expected status new, got %q", order.Status.Name)
	}
}
