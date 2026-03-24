package profiles

import "github.com/gofiber/fiber/v2"

type Handler struct {
	repo Repository
}

func RegisterRoutes(router fiber.Router, repo Repository) {
	h := &Handler{repo: repo}
	router.Post("/profiles", h.Create)

	router.Get("/profiles/:session_id", h.Get)
}

type createRequest struct {
	AgentID string `json:"agent_id"`
	Payload string `json:"payload"`
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req createRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	if req.AgentID == "" || req.Payload == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agent_id and payload required"})
	}

	userID := c.Locals("user_id").(int64)

	profile, err := h.repo.Create(userID, req.AgentID, req.Payload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save profile"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"session_id": profile.SessionID,
		"created_at": profile.CreatedAt,
	})
}

func (h *Handler) Get(c *fiber.Ctx) error {
	sessionID, err := c.ParamsInt("session_id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid session_id"})
	}

	profile, err := h.repo.Get(int64(sessionID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(profile)
}
