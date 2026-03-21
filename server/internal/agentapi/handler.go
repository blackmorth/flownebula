package agentapi

import (
	"flownebula/server/internal/sessions"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	sessions sessions.Repository
}

func RegisterRoutes(app fiber.Router, sessionRepo sessions.Repository) {
	h := &Handler{
		sessions: sessionRepo,
	}

	app.Post("/heartbeat", h.Heartbeat)
}

func (h *Handler) Heartbeat(c *fiber.Ctx) error {
	var req HeartbeatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid body",
		})
	}

	if req.AgentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "agent_id required",
		})
	}

	// L’agent est authentifié via JWT → user_id est dans Locals
	userID := c.Locals("user_id").(int64)

	// Créer une session pour cet agent
	session, err := h.sessions.Create(userID, req.AgentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create session",
		})
	}

	return c.JSON(HeartbeatResponse{
		Status:        "ok",
		SessionID:     session.ID,
		CheckInterval: 10, // secondes (configurable plus tard)
	})
}
