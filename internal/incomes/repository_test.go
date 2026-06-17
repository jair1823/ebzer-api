package incomes

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func newTestIncomeRepository(t *testing.T) (*sql.DB, Repository) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	schema := `
	CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		description TEXT NOT NULL,
		amount_charged REAL NOT NULL
	);

	CREATE TABLE income (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
		amount REAL NOT NULL,
		date TEXT NOT NULL DEFAULT (datetime('now')),
		deleted_at TEXT,
		deleted_by INTEGER,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	INSERT INTO orders (id, description, amount_charged)
	VALUES
		(1, 'First order', 25.50),
		(2, 'Second order', 80.00);
	`

	if _, err := db.Exec(schema); err != nil {
		if strings.Contains(err.Error(), "go-sqlite3 requires cgo") {
			t.Skipf("skipping SQLite repository test: %v", err)
		}
		t.Fatalf("create test schema: %v", err)
	}

	return db, NewRepository(db)
}

func TestRepositoryCreateUsesProvidedDate(t *testing.T) {
	db, repo := newTestIncomeRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	paymentDate := "2026-06-01"
	id, err := repo.Create(ctx, CreateIncomeDTO{
		OrderID: 1,
		Amount:  CustomFloat64(42.50),
		Date:    &paymentDate,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	var storedDate string
	if err := db.QueryRowContext(ctx, "SELECT date FROM income WHERE id = $1", id).Scan(&storedDate); err != nil {
		t.Fatalf("query stored date: %v", err)
	}
	if storedDate != paymentDate {
		t.Fatalf("expected date %q, got %q", paymentDate, storedDate)
	}
}

func TestRepositoryUpdateChangesDateAmountAndOrder(t *testing.T) {
	db, repo := newTestIncomeRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, "INSERT INTO income (id, order_id, amount, date) VALUES (1, 1, 10.00, '2026-05-01')"); err != nil {
		t.Fatalf("seed income: %v", err)
	}

	orderID := 2
	amount := CustomFloat64(60.75)
	paymentDate := "2026-06-02"
	if err := repo.Update(ctx, 1, UpdateIncomeDTO{
		OrderID: &orderID,
		Amount:  &amount,
		Date:    &paymentDate,
	}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	var storedOrderID int
	var storedAmount float64
	var storedDate string
	if err := db.QueryRowContext(ctx, "SELECT order_id, amount, date FROM income WHERE id = 1").Scan(&storedOrderID, &storedAmount, &storedDate); err != nil {
		t.Fatalf("query updated income: %v", err)
	}
	if storedOrderID != orderID {
		t.Fatalf("expected order_id %d, got %d", orderID, storedOrderID)
	}
	if storedAmount != float64(amount) {
		t.Fatalf("expected amount %.2f, got %.2f", amount, storedAmount)
	}
	if storedDate != paymentDate {
		t.Fatalf("expected date %q, got %q", paymentDate, storedDate)
	}
}

func TestRepositoryDeleteMissingReturnsIncomeNotFound(t *testing.T) {
	_, repo := newTestIncomeRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := repo.Delete(ctx, 999, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "income not found" {
		t.Fatalf("expected income not found, got %q", err.Error())
	}
}

func TestRepositoryDeleteSoftDeletesIncome(t *testing.T) {
	db, repo := newTestIncomeRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, "INSERT INTO income (id, order_id, amount) VALUES (1, 1, 10.00)"); err != nil {
		t.Fatalf("seed income: %v", err)
	}

	actorID := 3
	if err := repo.Delete(ctx, 1, &actorID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	income, err := repo.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if income != nil {
		t.Fatal("expected soft-deleted income to be hidden")
	}

	var deletedBy sql.NullInt64
	if err := db.QueryRowContext(ctx, "SELECT deleted_by FROM income WHERE id = 1").Scan(&deletedBy); err != nil {
		t.Fatalf("query deleted_by: %v", err)
	}
	if !deletedBy.Valid || deletedBy.Int64 != int64(actorID) {
		t.Fatalf("expected deleted_by %d, got %#v", actorID, deletedBy)
	}
}
