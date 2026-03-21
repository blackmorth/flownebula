package auth

import (
    "encoding/hex"
    "math/rand"
    "strconv"
    "time"
    "strings"

    "github.com/gofiber/fiber/v2"
    "golang.org/x/crypto/bcrypt"
)


type Handler struct {
	repo UserRepository
}

func NewHandler(repo UserRepository) *Handler {
    return &Handler{repo: repo}
}

func RegisterRoutes(app *fiber.App, repo UserRepository) {
	h := &Handler{repo: repo}

	g := app.Group("/auth")
	g.Post("/register", h.Register)
	g.Post("/login", h.Login)
	g.Get("/validate", h.Validate)
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email and password required"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}

	user := &User{
		Email:    req.Email,
		Password: string(hash),
	}

	if err := h.repo.Create(user); err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "user already exists"})
	}

	token, err := GenerateToken(user, 24*time.Hour)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
		},
		"token": token,
	})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	user, err := h.repo.FindByEmail(req.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token, err := GenerateToken(user, 24*time.Hour)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate token"})
	}

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
		},
		"token": token,
	})
}

func (h *Handler) Validate(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing Authorization header"})
	}

	// Format attendu : "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid Authorization header"})
	}

	tokenStr := parts[1]

	claims, err := ParseToken(tokenStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
	}

	return c.JSON(fiber.Map{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"exp":     claims.ExpiresAt,
	})
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
    users, err := h.repo.FindAll()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to list users"})
    }
    return c.JSON(users)
}

func (h *Handler) EnableAgent(c *fiber.Ctx) error {
    id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
    if err := h.repo.UpdateAgentEnabled(id, true); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to enable agent"})
    }
    return c.JSON(fiber.Map{"status": "enabled"})
}

func (h *Handler) DisableAgent(c *fiber.Ctx) error {
    id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
    if err := h.repo.UpdateAgentEnabled(id, false); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to disable agent"})
    }
    return c.JSON(fiber.Map{"status": "disabled"})
}

func (h *Handler) RegenerateAgentToken(c *fiber.Ctx) error {
    id, _ := strconv.ParseInt(c.Params("id"), 10, 64)

    tokenBytes := make([]byte, 32)
    rand.Read(tokenBytes)
    newToken := "nebula_" + hex.EncodeToString(tokenBytes)

    if err := h.repo.UpdateAgentToken(id, newToken); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to regenerate token"})
    }

    return c.JSON(fiber.Map{"agent_token": newToken})
}

func RegisterAdminRoutes(router fiber.Router, repo UserRepository) {
    h := NewHandler(repo)

    router.Get("/users", h.ListUsers)
    router.Post("/users/:id/agent/enable", h.EnableAgent)
    router.Post("/users/:id/agent/disable", h.DisableAgent)
    router.Post("/users/:id/agent/regenerate", h.RegenerateAgentToken)
}

