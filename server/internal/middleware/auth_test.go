package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"flownebula/server/internal/auth"

	"github.com/gofiber/fiber/v2"
)

func TestJWTProtectedRejectsMissingAuthorizationHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/protected", JWTProtected(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestJWTProtectedAllowsValidBearerToken(t *testing.T) {
	app := fiber.New()
	app.Get("/protected", JWTProtected(), func(c *fiber.Ctx) error {
		uid := c.Locals("user_id").(int64)
		if uid != 99 {
			t.Fatalf("expected user_id=99, got %d", uid)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	token, err := auth.GenerateToken(&auth.User{ID: 99, Email: "dev@example.com", Roles: []string{"ROLE_ADMIN"}}, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
}

func TestRequireRoleRejectsUnauthorizedRole(t *testing.T) {
	app := fiber.New()
	app.Get("/admin", func(c *fiber.Ctx) error {
		c.Locals("roles", []string{"ROLE_USER"})
		return c.Next()
	}, RequireRole("ROLE_ADMIN"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}

	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if payload["error"] != "forbidden" {
		t.Fatalf("unexpected error payload: %#v", payload)
	}
}
