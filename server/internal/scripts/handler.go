package scripts

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"flownebula/server/internal/sessions"

	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
)

type Handler struct {
	sessionRepo sessions.Repository
	rootDir     string
}

type runScriptRequest struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
}

func RegisterRoutes(router fiber.Router, sessionRepo sessions.Repository) {
	root := os.Getenv("PHP_SCRIPTS_ROOT")
	if root == "" {
		root = "."
	}
	if abs, err := filepath.Abs(root); err == nil {
		root = abs
	}

	h := &Handler{sessionRepo: sessionRepo, rootDir: root}
	router.Post("/run", h.Run)
}

func (h *Handler) Run(c *fiber.Ctx) error {
	var req runScriptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Path == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "path required"})
	}

	scriptPath, err := h.resolveScriptPath(req.Path)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if _, err := os.Stat(scriptPath); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "script not found"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "php", append([]string{scriptPath}, req.Args...)...)
	cmd.Dir = h.rootDir
	output, execErr := cmd.CombinedOutput()

	exitCode := 0
	errorText := ""
	if execErr != nil {
		errorText = execErr.Error()
		var exitErr *exec.ExitError
		if errors.As(execErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	payloadMap := map[string]any{
		"type":        "php_script_execution",
		"script_path": scriptPath,
		"args":        req.Args,
		"output":      string(output),
		"exit_code":   exitCode,
		"error":       errorText,
		"executed_at": time.Now().UTC().Format(time.RFC3339),
	}
	payloadRaw, _ := json.Marshal(payloadMap)

	userID := c.Locals("user_id").(int64)
	session := &sessions.Session{
		UserID:    userID,
		AgentID:   "local-php",
		Payload:   datatypes.JSON(payloadRaw),
		CreatedAt: time.Now(),
	}
	if err := h.sessionRepo.Create(session); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save session"})
	}

	return c.JSON(fiber.Map{"session": session, "result": payloadMap})
}

func (h *Handler) resolveScriptPath(raw string) (string, error) {
	clean := filepath.Clean(raw)
	if !strings.HasSuffix(strings.ToLower(clean), ".php") {
		return "", errors.New("only .php scripts are allowed")
	}

	candidate := clean
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(h.rootDir, candidate)
	}
	abs, err := filepath.Abs(candidate)
	if err != nil {
		return "", errors.New("invalid script path")
	}

	rootWithSep := h.rootDir + string(os.PathSeparator)
	if abs != h.rootDir && !strings.HasPrefix(abs, rootWithSep) {
		return "", errors.New("script path must be inside scripts root")
	}

	return abs, nil
}
