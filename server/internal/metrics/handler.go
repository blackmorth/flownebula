package metrics

import (
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	repo Repository
}

func RegisterRoutes(app fiber.Router, repo Repository) {
	h := &Handler{repo: repo}
	app.Post("/", h.Push)
}

type MetricRequest struct {
	SessionID    int64   `json:"session_id"`
	CPUUsage     float64 `json:"cpu_usage"`
	RAMUsage     float64 `json:"ram_usage"`
	LoadAvg      float64 `json:"load_avg"`
	ProcessCount int     `json:"process_count"`
}

func (h *Handler) Push(c *fiber.Ctx) error {
	var req MetricRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	if req.SessionID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "session_id required"})
	}

	m := &Metric{
		SessionID:    req.SessionID,
		CPUUsage:     req.CPUUsage,
		RAMUsage:     req.RAMUsage,
		LoadAvg:      req.LoadAvg,
		ProcessCount: req.ProcessCount,
	}

	if err := h.repo.Insert(m); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to insert metric"})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}
