package agentapi

import (
	"encoding/json"
	"flownebula/server/internal/sessions"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

type Handler struct {
	sessions sessions.Repository
}

func RegisterRoutes(app fiber.Router, sessionRepo sessions.Repository, uploadMiddlewares ...fiber.Handler) {
	h := &Handler{
		sessions: sessionRepo,
	}

	app.Post("/heartbeat", h.Heartbeat)
	app.Post("/session-upload", append(uploadMiddlewares, h.UploadSession)...)
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
	//userID := c.Locals("user_id").(int64)

	// TODO: mettre à jour l’état de l’agent (plus tard)
	// h.agents.UpdateHeartbeat(userID, req.AgentID, req.Version)

	return c.JSON(HeartbeatResponse{
		Status:        "ok",
		SessionID:     0, // plus utilisé
		CheckInterval: 10,
	})
}

func (h *Handler) UploadSession(c *fiber.Ctx) error {
	var userID int64 = 0
	if v := c.Locals("user_id"); v != nil {
		if id, ok := v.(int64); ok {
			userID = id
		}
	}

	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		log.Printf("[UPLOAD] invalid JSON body: %v", err)
		return fiber.ErrBadRequest
	}

	agentSessionID := ""
	if v, ok := payload["agent_session_id"].(string); ok {
		agentSessionID = v
	}

	agentID := ""
	if v, ok := payload["agent_id"].(string); ok {
		agentID = v
	}
	service := ""
	if v, ok := payload["service"].(string); ok {
		service = v
	}
	endpoint := ""
	if v, ok := payload["endpoint"].(string); ok {
		endpoint = v
	}
	release := ""
	if v, ok := payload["release"].(string); ok {
		release = v
	}
	tags := ""
	if rawTags, ok := payload["tags"]; ok {
		if b, err := json.Marshal(rawTags); err == nil {
			tags = string(b)
		}
	}

	jsonBytes, _ := json.Marshal(payload)

	session := &sessions.Session{
		UserID:         userID,
		AgentID:        agentID,
		AgentSessionID: agentSessionID,
		Service:        service,
		Endpoint:       endpoint,
		Release:        release,
		Tags:           tags,
		Payload:        datatypes.JSON(jsonBytes),
		CreatedAt:      time.Now(),
	}

	if err := h.sessions.Create(session); err != nil {
		log.Printf("[UPLOAD] DB error: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to store session",
		})
	}

	log.Printf("[UPLOAD] stored session=%d user_id=%d agent_id=%s payload_bytes=%d", session.ID, userID, agentID, len(jsonBytes))

	return c.Status(201).JSON(fiber.Map{
		"session_id": session.ID,
	})
}
