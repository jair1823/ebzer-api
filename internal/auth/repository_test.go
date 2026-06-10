package auth

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestRepository(t *testing.T) (*sql.DB, Repository) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'operator' CHECK(role IN ('admin', 'operator', 'guest')),
			is_active INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		);
	`)
	if err != nil {
		if strings.Contains(err.Error(), "go-sqlite3 requires cgo") {
			t.Skipf("skipping SQLite repository test: %v", err)
		}
		t.Fatalf("create schema: %v", err)
	}

	return db, NewRepository(db)
}

func TestRepositoryCreateListUpdateAndDeactivate(t *testing.T) {
	_, repo := newTestRepository(t)
	ctx := context.Background()

	id, err := repo.Create(ctx, CreateUserRequest{
		Name:     "Owner",
		Username: "OWNER",
		Email:    "OWNER@EXAMPLE.COM",
		Role:     RoleAdmin,
	}, "hash")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	user, err := repo.GetByEmail(ctx, "owner@example.com")
	if err != nil {
		t.Fatalf("GetByEmail returned error: %v", err)
	}
	if user == nil || user.ID != id || user.Username != "owner" || user.Email != "owner@example.com" || user.Role != RoleAdmin {
		t.Fatalf("unexpected created user: %#v", user)
	}
	user, err = repo.GetByUsername(ctx, "OWNER")
	if err != nil {
		t.Fatalf("GetByUsername returned error: %v", err)
	}
	if user == nil || user.ID != id {
		t.Fatalf("expected to find user by username, got %#v", user)
	}
	if !user.IsActive {
		t.Fatal("expected user to be active by default")
	}

	newRole := RoleGuest
	newName := "Guest Owner"
	newUsername := "guest-owner"
	if err := repo.Update(ctx, id, UpdateUserRequest{
		Name:     &newName,
		Username: &newUsername,
		Role:     &newRole,
	}, nil); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(users) != 1 || users[0].Name != newName || users[0].Username != newUsername || users[0].Role != RoleGuest {
		t.Fatalf("unexpected users list: %#v", users)
	}

	if err := repo.Deactivate(ctx, id); err != nil {
		t.Fatalf("Deactivate returned error: %v", err)
	}
	user, err = repo.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if user == nil || user.IsActive {
		t.Fatalf("expected deactivated user, got %#v", user)
	}
}

func TestRepositoryRejectsDuplicateUsername(t *testing.T) {
	_, repo := newTestRepository(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, CreateUserRequest{
		Name:     "Owner",
		Username: "owner",
		Email:    "owner@example.com",
		Role:     RoleAdmin,
	}, "hash")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if _, err := repo.Create(ctx, CreateUserRequest{
		Name:     "Other Owner",
		Username: "OWNER",
		Email:    "other@example.com",
		Role:     RoleOperator,
	}, "hash"); err == nil {
		t.Fatal("expected duplicate username to fail")
	}
}
