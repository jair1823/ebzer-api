package agenda

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const agendaTestSchema = `
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
	paid_at TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE agenda_items (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type TEXT NOT NULL DEFAULT 'note' CHECK(type IN ('note', 'task', 'reminder')),
	title TEXT NOT NULL,
	content TEXT,
	status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'done', 'archived')),
	priority TEXT NOT NULL DEFAULT 'medium' CHECK(priority IN ('low', 'medium', 'high')),
	due_date TEXT,
	completed_at TEXT,
	order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO order_statuses (id, name, display_name, color, order_position, is_system_status)
VALUES (1, 'new', 'New', '#3B82F6', 1, 1);

INSERT INTO orders (id, description, amount_charged, status_id, client_name, client_phone, estimated_delivery_date)
VALUES (1, 'Print shirts', 150.00, 1, 'Ana López', '555-1234', '2026-06-30');
`

func newTestAgendaRepository(t *testing.T) (*sql.DB, Repository) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec(agendaTestSchema); err != nil {
		if strings.Contains(err.Error(), "go-sqlite3 requires cgo") {
			t.Skipf("skipping SQLite repository test: %v", err)
		}
		t.Fatalf("create test schema: %v", err)
	}

	return db, NewRepository(db)
}

func ctx1s() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Second)
}

// -------------------- Create --------------------

func TestRepositoryCreate(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	id, err := repo.Create(ctx, CreateAgendaItemDTO{
		Type:     TypeTask,
		Title:    "Buy fabric",
		Priority: PriorityHigh,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}
}

// -------------------- GetByID --------------------

func TestRepositoryGetByID_ReturnsNilForMissing(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	item, err := repo.GetByID(ctx, 999)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if item != nil {
		t.Fatal("expected nil for missing item")
	}
}

func TestRepositoryGetByID_PopulatesOrderSummary(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	orderID := 1
	id, err := repo.Create(ctx, CreateAgendaItemDTO{
		Title:   "Check order",
		Type:    TypeNote,
		OrderID: &orderID,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	item, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if item == nil {
		t.Fatal("expected item")
	}
	if item.Order == nil {
		t.Fatal("expected order summary to be populated")
	}
	if item.Order.Description != "Print shirts" {
		t.Fatalf("expected description 'Print shirts', got %q", item.Order.Description)
	}
	if item.Order.ClientName == nil || *item.Order.ClientName != "Ana López" {
		t.Fatalf("expected client_name 'Ana López', got %v", item.Order.ClientName)
	}
}

func TestRepositoryGetByID_NilOrderSummaryWhenNoOrder(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	id, err := repo.Create(ctx, CreateAgendaItemDTO{Title: "Plain note", Type: TypeNote})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	item, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if item == nil {
		t.Fatal("expected item")
	}
	if item.Order != nil {
		t.Fatalf("expected nil order summary, got %+v", item.Order)
	}
}

// -------------------- GetAll filters --------------------

func TestRepositoryGetAll_DefaultReturnsItems(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	repo.Create(ctx, CreateAgendaItemDTO{Title: "A", Type: TypeNote})
	repo.Create(ctx, CreateAgendaItemDTO{Title: "B", Type: TypeTask})

	items, err := repo.GetAll(ctx, FilterAgendaItemsDTO{Status: "pending"})
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestRepositoryGetAll_FilterByStatus(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	id1, _ := repo.Create(ctx, CreateAgendaItemDTO{Title: "Pending", Type: TypeTask})
	id2, _ := repo.Create(ctx, CreateAgendaItemDTO{Title: "Done soon", Type: TypeTask})
	repo.Complete(ctx, id2)
	_ = id1

	pending, err := repo.GetAll(ctx, FilterAgendaItemsDTO{Status: "pending"})
	if err != nil {
		t.Fatalf("GetAll pending: %v", err)
	}
	if len(pending) != 1 || pending[0].Title != "Pending" {
		t.Fatalf("expected 1 pending item, got %d", len(pending))
	}

	done, err := repo.GetAll(ctx, FilterAgendaItemsDTO{Status: "done"})
	if err != nil {
		t.Fatalf("GetAll done: %v", err)
	}
	if len(done) != 1 || done[0].Title != "Done soon" {
		t.Fatalf("expected 1 done item, got %d", len(done))
	}
}

func TestRepositoryGetAll_AllStatusReturnsEverything(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	id, _ := repo.Create(ctx, CreateAgendaItemDTO{Title: "Item A", Type: TypeNote})
	repo.Archive(ctx, id)
	repo.Create(ctx, CreateAgendaItemDTO{Title: "Item B", Type: TypeTask})

	all, err := repo.GetAll(ctx, FilterAgendaItemsDTO{})
	if err != nil {
		t.Fatalf("GetAll all: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 total items, got %d", len(all))
	}
}

func TestRepositoryGetAll_FilterByType(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	repo.Create(ctx, CreateAgendaItemDTO{Title: "Note", Type: TypeNote})
	repo.Create(ctx, CreateAgendaItemDTO{Title: "Task", Type: TypeTask})

	notes, err := repo.GetAll(ctx, FilterAgendaItemsDTO{Status: "pending", Type: string(TypeNote)})
	if err != nil {
		t.Fatalf("GetAll type filter: %v", err)
	}
	if len(notes) != 1 || notes[0].Type != TypeNote {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}
}

func TestRepositoryGetAll_FilterByPriority(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	repo.Create(ctx, CreateAgendaItemDTO{Title: "High prio", Type: TypeTask, Priority: PriorityHigh})
	repo.Create(ctx, CreateAgendaItemDTO{Title: "Low prio", Type: TypeTask, Priority: PriorityLow})

	high, err := repo.GetAll(ctx, FilterAgendaItemsDTO{Status: "pending", Priority: string(PriorityHigh)})
	if err != nil {
		t.Fatalf("GetAll priority filter: %v", err)
	}
	if len(high) != 1 || high[0].Priority != PriorityHigh {
		t.Fatalf("expected 1 high priority item, got %d", len(high))
	}
}

func TestRepositoryGetAll_FilterByOrderID(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	orderID := 1
	repo.Create(ctx, CreateAgendaItemDTO{Title: "With order", Type: TypeNote, OrderID: &orderID})
	repo.Create(ctx, CreateAgendaItemDTO{Title: "Without order", Type: TypeNote})

	filtered, err := repo.GetAll(ctx, FilterAgendaItemsDTO{Status: "pending", OrderID: &orderID})
	if err != nil {
		t.Fatalf("GetAll order_id filter: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Title != "With order" {
		t.Fatalf("expected 1 item linked to order, got %d", len(filtered))
	}
}

func TestRepositoryGetAll_FilterByDueDateRange(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	past := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	future := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	repo.Create(ctx, CreateAgendaItemDTO{Title: "Past task", Type: TypeTask, DueDate: &past})
	repo.Create(ctx, CreateAgendaItemDTO{Title: "Future task", Type: TypeTask, DueDate: &future})

	from := "2026-06-01"
	items, err := repo.GetAll(ctx, FilterAgendaItemsDTO{Status: "pending", From: &from})
	if err != nil {
		t.Fatalf("GetAll from filter: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Future task" {
		t.Fatalf("expected 1 future task, got %d", len(items))
	}
}

func TestRepositoryGetAll_SearchAcrossTitleAndContent(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	content := "fabric samples"
	repo.Create(ctx, CreateAgendaItemDTO{Title: "Buy fabric", Type: TypeNote, Content: &content})
	repo.Create(ctx, CreateAgendaItemDTO{Title: "Call client", Type: TypeTask})

	search := "fabric"
	items, err := repo.GetAll(ctx, FilterAgendaItemsDTO{Status: "pending", Search: &search})
	if err != nil {
		t.Fatalf("GetAll search: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Buy fabric" {
		t.Fatalf("expected 1 search result, got %d", len(items))
	}
}

func TestRepositoryGetAll_SearchMatchesLinkedOrderDescription(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	orderID := 1
	repo.Create(ctx, CreateAgendaItemDTO{Title: "Follow-up", Type: TypeNote, OrderID: &orderID})
	repo.Create(ctx, CreateAgendaItemDTO{Title: "Other item", Type: TypeNote})

	// orders.description = "Print shirts"
	search := "Print shirts"
	items, err := repo.GetAll(ctx, FilterAgendaItemsDTO{Status: "pending", Search: &search})
	if err != nil {
		t.Fatalf("GetAll search by order description: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Follow-up" {
		t.Fatalf("expected 1 result matching order description, got %d", len(items))
	}
}

// -------------------- Update --------------------

func TestRepositoryUpdate(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	id, _ := repo.Create(ctx, CreateAgendaItemDTO{Title: "Old title", Type: TypeNote})
	newTitle := "New title"
	if err := repo.Update(ctx, id, UpdateAgendaItemDTO{Title: &newTitle}); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	item, _ := repo.GetByID(ctx, id)
	if item.Title != "New title" {
		t.Fatalf("expected title 'New title', got %q", item.Title)
	}
}

func TestRepositoryUpdate_NotFound(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	title := "x"
	err := repo.Update(ctx, 999, UpdateAgendaItemDTO{Title: &title})
	if err == nil || err.Error() != "agenda item not found" {
		t.Fatalf("expected not found error, got %v", err)
	}
}

// -------------------- Delete --------------------

func TestRepositoryDelete(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	id, _ := repo.Create(ctx, CreateAgendaItemDTO{Title: "Delete me", Type: TypeNote})
	if err := repo.Delete(ctx, id); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	item, _ := repo.GetByID(ctx, id)
	if item != nil {
		t.Fatal("expected item to be deleted")
	}
}

func TestRepositoryDelete_NotFound(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	err := repo.Delete(ctx, 999)
	if err == nil || err.Error() != "agenda item not found" {
		t.Fatalf("expected not found error, got %v", err)
	}
}

// -------------------- Complete --------------------

func TestRepositoryComplete_SetsStatusAndCompletedAt(t *testing.T) {
	db, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	id, _ := repo.Create(ctx, CreateAgendaItemDTO{Title: "Finish this", Type: TypeTask})
	if err := repo.Complete(ctx, id); err != nil {
		t.Fatalf("Complete returned error: %v", err)
	}

	var status, completedAt sql.NullString
	db.QueryRowContext(ctx, "SELECT status, completed_at FROM agenda_items WHERE id = $1", id).
		Scan(&status, &completedAt)

	if status.String != "done" {
		t.Fatalf("expected status 'done', got %q", status.String)
	}
	if !completedAt.Valid || completedAt.String == "" {
		t.Fatal("expected completed_at to be set")
	}
}

func TestRepositoryComplete_NotFound(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	err := repo.Complete(ctx, 999)
	if err == nil || err.Error() != "agenda item not found" {
		t.Fatalf("expected not found error, got %v", err)
	}
}

// -------------------- Archive --------------------

func TestRepositoryArchive_SetsStatusAndPreservesCompletedAt(t *testing.T) {
	db, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	id, _ := repo.Create(ctx, CreateAgendaItemDTO{Title: "Old task", Type: TypeTask})

	// Complete first to set completed_at, then archive
	repo.Complete(ctx, id)

	var completedAtBefore sql.NullString
	db.QueryRowContext(ctx, "SELECT completed_at FROM agenda_items WHERE id = $1", id).Scan(&completedAtBefore)

	if err := repo.Archive(ctx, id); err != nil {
		t.Fatalf("Archive returned error: %v", err)
	}

	var status, completedAtAfter sql.NullString
	db.QueryRowContext(ctx, "SELECT status, completed_at FROM agenda_items WHERE id = $1", id).
		Scan(&status, &completedAtAfter)

	if status.String != "archived" {
		t.Fatalf("expected status 'archived', got %q", status.String)
	}
	if completedAtBefore.String != completedAtAfter.String {
		t.Fatalf("expected completed_at unchanged: before=%q after=%q",
			completedAtBefore.String, completedAtAfter.String)
	}
}

func TestRepositoryArchive_NotFound(t *testing.T) {
	_, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	err := repo.Archive(ctx, 999)
	if err == nil || err.Error() != "agenda item not found" {
		t.Fatalf("expected not found error, got %v", err)
	}
}

// -------------------- Order SET NULL on delete --------------------

func TestRepositoryOrderIDSetNullOnOrderDelete(t *testing.T) {
	db, repo := newTestAgendaRepository(t)
	ctx, cancel := ctx1s()
	defer cancel()

	orderID := 1
	id, _ := repo.Create(ctx, CreateAgendaItemDTO{Title: "Linked", Type: TypeNote, OrderID: &orderID})

	db.ExecContext(ctx, "DELETE FROM orders WHERE id = 1")

	item, err := repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID after order delete: %v", err)
	}
	if item.OrderID != nil {
		t.Fatalf("expected order_id to be null after order delete, got %v", item.OrderID)
	}
	if item.Order != nil {
		t.Fatalf("expected nil order summary after order delete")
	}
}
