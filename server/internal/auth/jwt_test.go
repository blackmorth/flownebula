package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	user := &User{ID: 7, Email: "qa@example.com", Roles: []string{"ROLE_USER", "ROLE_ADMIN"}}
	token, err := GenerateToken(user, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.UserID != user.ID {
		t.Fatalf("expected user id %d, got %d", user.ID, claims.UserID)
	}
	if claims.Email != user.Email {
		t.Fatalf("expected email %q, got %q", user.Email, claims.Email)
	}
	if len(claims.Roles) != len(user.Roles) {
		t.Fatalf("expected %d roles, got %d", len(user.Roles), len(claims.Roles))
	}
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	user := &User{ID: 9, Email: "expired@example.com", Roles: []string{"ROLE_USER"}}
	token, err := GenerateToken(user, -time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if _, err := ParseToken(token); err == nil {
		t.Fatalf("expected ParseToken to fail for expired token")
	}
}
