package auth

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestRequireAuthAndRole(t *testing.T) {
	repo := newFakeRepository()
	jwt := NewJWTService("test-secret")
	user := &User{
		ID:       1,
		Name:     "Operator",
		Username: "operator",
		Email:    "operator@example.com",
		Role:     RoleOperator,
		IsActive: true,
	}
	repo.users[user.ID] = user

	token, err := jwt.Generate(user, TokenTypeAccess, time.Minute)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	app := fiber.New()
	app.Use(RequireAuth(jwt, repo))
	app.Get("/read", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})
	app.Post("/admin", RequireRole(RoleAdmin), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/read", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", res.StatusCode)
	}

	req = httptest.NewRequest("GET", "/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err = app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if res.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204 with valid token, got %d", res.StatusCode)
	}

	req = httptest.NewRequest("POST", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err = app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if res.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403 for non-admin, got %d", res.StatusCode)
	}

	user.Username = "changed-operator"
	req = httptest.NewRequest("GET", "/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err = app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for changed username claim, got %d", res.StatusCode)
	}

	user.Username = "operator"
	user.IsActive = false
	req = httptest.NewRequest("GET", "/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err = app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if res.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for inactive user, got %d", res.StatusCode)
	}
}

var _ Repository = (*fakeRepository)(nil)
