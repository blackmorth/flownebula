package sessions

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	repo Repository
}

func RegisterRoutes(app fiber.Router, repo Repository) {
	h := &Handler{repo: repo}

	app.Post("/", h.Create)
	app.Get("/", h.List)
	app.Get("/:id", h.Get)
}

type createRequest struct {
	AgentID string `json:"agent_id"`
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req createRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	if req.AgentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agent_id required"})
	}

	userID := c.Locals("user_id").(int64)

	session, err := h.repo.Create(userID, req.AgentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create session"})
	}

	return c.Status(fiber.StatusCreated).JSON(session)
}

func (h *Handler) List(c *fiber.Ctx) error {
	sessions, err := h.repo.List()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list sessions"})
	}

	return c.JSON(sessions)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	session, err := h.repo.Get(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "session not found"})
	}

	return c.JSON(session)
}
