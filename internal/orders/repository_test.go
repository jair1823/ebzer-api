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
		platform TEXT NOT NULL DEFAULT 'whatsapp' CHECK(platform IN ('whatsapp', 'instagram', 'facebook')),
		client_name TEXT,
		client_phone TEXT,
		notes TEXT,
		paid_at TEXT,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE income (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
		amount REAL NOT NULL,
		date TEXT NOT NULL DEFAULT (datetime('now')),
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
	if first.Platform != PlatformWhatsApp {
		t.Fatalf("expected First order platform whatsapp, got %q", first.Platform)
	}

	second := byDescription["Second order"]
	if second.Status == nil {
		t.Fatal("expected Second order to include status")
	}
	if second.Status.Name != "completed" {
		t.Fatalf("expected Second order status completed, got %q", second.Status.Name)
	}
	if second.Platform != PlatformWhatsApp {
		t.Fatalf("expected Second order platform whatsapp, got %q", second.Platform)
	}
}

func TestRepositoryGetAllPopulatesPaymentStatus(t *testing.T) {
	db, repo := newTestOrderRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, "INSERT INTO income (order_id, amount) VALUES (1, 10.00)"); err != nil {
		t.Fatalf("seed income: %v", err)
	}

	orders, err := repo.GetAll(ctx, nil, nil, nil)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}

	byID := map[int]Order{}
	for _, order := range orders {
		byID[order.ID] = order
	}

	first := byID[1]
	if first.PaymentStatus == nil {
		t.Fatal("expected payment status")
	}
	if first.PaymentStatus.TotalPaid != 10.00 {
		t.Fatalf("expected total paid 10.00, got %.2f", first.PaymentStatus.TotalPaid)
	}
	if first.PaymentStatus.AmountCharged != 25.50 {
		t.Fatalf("expected amount charged 25.50, got %.2f", first.PaymentStatus.AmountCharged)
	}
	if first.PaymentStatus.Remaining != 15.50 {
		t.Fatalf("expected remaining 15.50, got %.2f", first.PaymentStatus.Remaining)
	}
	if first.PaymentStatus.PercentagePaid == 0 {
		t.Fatal("expected payment percentage greater than 0")
	}
	if first.PaymentStatus.IsFullyPaid {
		t.Fatal("expected order not to be fully paid")
	}

	second := byID[2]
	if second.PaymentStatus == nil {
		t.Fatal("expected payment status for unpaid order")
	}
	if second.PaymentStatus.TotalPaid != 0 {
		t.Fatalf("expected unpaid order total paid 0, got %.2f", second.PaymentStatus.TotalPaid)
	}
	if second.PaymentStatus.Remaining != 80.00 {
		t.Fatalf("expected unpaid order remaining 80.00, got %.2f", second.PaymentStatus.Remaining)
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
	if order.PaymentStatus == nil {
		t.Fatal("expected payment status")
	}
	if order.PaymentStatus.AmountCharged != 25.50 {
		t.Fatalf("expected amount charged 25.50, got %.2f", order.PaymentStatus.AmountCharged)
	}
	if order.Platform != PlatformWhatsApp {
		t.Fatalf("expected platform whatsapp, got %q", order.Platform)
	}
}

func TestServiceCreateDefaultsPlatformToWhatsApp(t *testing.T) {
	_, repo := newTestOrderRepository(t)
	svc := NewService(repo)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	id, err := svc.Create(ctx, CreateOrderDTO{
		Description:   "Default platform",
		AmountCharged: CustomFloat64(55.00),
		DeliveryType:  DeliveryPickup,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	order, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if order == nil {
		t.Fatal("expected created order")
	}
	if order.Platform != PlatformWhatsApp {
		t.Fatalf("expected platform whatsapp, got %q", order.Platform)
	}
}

func TestRepositoryCreateWithPlatform(t *testing.T) {
	_, repo := newTestOrderRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	id, err := repo.Create(ctx, CreateOrderDTO{
		Description:   "Instagram order",
		AmountCharged: CustomFloat64(42.00),
		StatusID:      1,
		DeliveryType:  DeliveryPickup,
		Platform:      PlatformInstagram,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	order, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if order == nil {
		t.Fatal("expected created order")
	}
	if order.Platform != PlatformInstagram {
		t.Fatalf("expected platform instagram, got %q", order.Platform)
	}
}

func TestRepositoryUpdatePlatform(t *testing.T) {
	_, repo := newTestOrderRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	platform := PlatformFacebook
	if err := repo.Update(ctx, 1, UpdateOrderDTO{Platform: &platform}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	order, err := repo.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if order == nil {
		t.Fatal("expected order")
	}
	if order.Platform != PlatformFacebook {
		t.Fatalf("expected platform facebook, got %q", order.Platform)
	}
}

func TestRepositoryFinishOrderCreatesIncomeForFullAmount(t *testing.T) {
	db, repo := newTestOrderRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result, err := repo.FinishOrder(ctx, 1)
	if err != nil {
		t.Fatalf("FinishOrder returned error: %v", err)
	}
	if result == nil || !result.Finished {
		t.Fatalf("expected finished result, got %#v", result)
	}
	if !result.IncomeCreated {
		t.Fatal("expected income to be created")
	}
	if result.IncomeID == nil {
		t.Fatal("expected income id")
	}
	if result.AmountPaid != 25.50 {
		t.Fatalf("expected amount paid 25.50, got %.2f", result.AmountPaid)
	}
	if result.TotalPaid != 25.50 {
		t.Fatalf("expected total paid 25.50, got %.2f", result.TotalPaid)
	}
	if result.Remaining != 0 {
		t.Fatalf("expected remaining 0, got %.2f", result.Remaining)
	}
	if !result.IsFullyPaid {
		t.Fatal("expected order to be fully paid")
	}
	if result.PaidAt == nil || !result.PaidAt.Valid {
		t.Fatalf("expected paid_at to be set, got %#v", result.PaidAt)
	}

	var statusName string
	if err := db.QueryRowContext(ctx, `
		SELECT s.name
		FROM orders o
		JOIN order_statuses s ON s.id = o.status_id
		WHERE o.id = 1
	`).Scan(&statusName); err != nil {
		t.Fatalf("query status: %v", err)
	}
	if statusName != "completed" {
		t.Fatalf("expected completed status, got %q", statusName)
	}

	var incomeAmount float64
	if err := db.QueryRowContext(ctx, "SELECT amount FROM income WHERE order_id = 1").Scan(&incomeAmount); err != nil {
		t.Fatalf("query income: %v", err)
	}
	if incomeAmount != 25.50 {
		t.Fatalf("expected income amount 25.50, got %.2f", incomeAmount)
	}
}

func TestRepositoryFinishOrderCreatesIncomeForPendingAmount(t *testing.T) {
	db, repo := newTestOrderRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, "INSERT INTO income (order_id, amount) VALUES (1, 10.00)"); err != nil {
		t.Fatalf("seed partial income: %v", err)
	}

	result, err := repo.FinishOrder(ctx, 1)
	if err != nil {
		t.Fatalf("FinishOrder returned error: %v", err)
	}
	if !result.IncomeCreated {
		t.Fatal("expected pending income to be created")
	}
	if result.AmountPaid != 15.50 {
		t.Fatalf("expected amount paid 15.50, got %.2f", result.AmountPaid)
	}
	if result.TotalPaid != 25.50 {
		t.Fatalf("expected total paid 25.50, got %.2f", result.TotalPaid)
	}

	var incomeCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM income WHERE order_id = 1").Scan(&incomeCount); err != nil {
		t.Fatalf("query income count: %v", err)
	}
	if incomeCount != 2 {
		t.Fatalf("expected 2 income records, got %d", incomeCount)
	}
}

func TestRepositoryFinishOrderDoesNotCreateIncomeWhenAlreadyPaid(t *testing.T) {
	db, repo := newTestOrderRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, "INSERT INTO income (order_id, amount) VALUES (1, 30.00)"); err != nil {
		t.Fatalf("seed paid income: %v", err)
	}

	result, err := repo.FinishOrder(ctx, 1)
	if err != nil {
		t.Fatalf("FinishOrder returned error: %v", err)
	}
	if result.IncomeCreated {
		t.Fatal("expected no income to be created")
	}
	if result.IncomeID != nil {
		t.Fatalf("expected nil income id, got %#v", result.IncomeID)
	}
	if result.AmountPaid != 0 {
		t.Fatalf("expected amount paid 0, got %.2f", result.AmountPaid)
	}
	if result.TotalPaid != 30.00 {
		t.Fatalf("expected total paid 30.00, got %.2f", result.TotalPaid)
	}
	if result.Remaining != 0 {
		t.Fatalf("expected remaining 0, got %.2f", result.Remaining)
	}
	if !result.IsFullyPaid {
		t.Fatal("expected order to be fully paid")
	}

	var incomeCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM income WHERE order_id = 1").Scan(&incomeCount); err != nil {
		t.Fatalf("query income count: %v", err)
	}
	if incomeCount != 1 {
		t.Fatalf("expected 1 income record, got %d", incomeCount)
	}
}

func TestRepositoryUpdateStatusDoesNotCreateIncomeOrPaidAt(t *testing.T) {
	db, repo := newTestOrderRepository(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	completedStatusID := 2
	if err := repo.Update(ctx, 1, UpdateOrderDTO{StatusID: &completedStatusID}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	var incomeCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM income WHERE order_id = 1").Scan(&incomeCount); err != nil {
		t.Fatalf("query income count: %v", err)
	}
	if incomeCount != 0 {
		t.Fatalf("expected no income records, got %d", incomeCount)
	}

	var paidAt sql.NullString
	if err := db.QueryRowContext(ctx, "SELECT paid_at FROM orders WHERE id = 1").Scan(&paidAt); err != nil {
		t.Fatalf("query paid_at: %v", err)
	}
	if paidAt.Valid {
		t.Fatalf("expected paid_at to remain null, got %q", paidAt.String)
	}
}
