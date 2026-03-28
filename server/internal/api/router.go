package api

import (
	"flownebula/server/internal/agentapi"
	"flownebula/server/internal/auth"
	"flownebula/server/internal/db"
	"flownebula/server/internal/middleware"
	"flownebula/server/internal/profiles"
	"flownebula/server/internal/sessions"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"log"
)

func New() *fiber.App {
	cfg := LoadConfig()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type, Authorization",
		AllowCredentials: true,
	}))

	app.Options("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	sqlite := db.Open("nebula.db")
	db.Migrate(sqlite)
	deleted, err := db.ApplyRetention(sqlite, cfg.SessionRetentionInDays)
	if err != nil {
		log.Printf("failed to apply session retention policy: %v", err)
	} else if deleted > 0 {
		log.Printf("retention cleanup removed %d old sessions", deleted)
	}

	authRepo := auth.NewSQLiteRepo(sqlite)
	sessionRepo := sessions.NewSQLiteRepo(sqlite)
	profilesRepo := profiles.NewSQLiteRepo(sqlite)

	auth.RegisterRoutes(app, authRepo)

	// Middleware appliqué ici (hors du package auth)
	authProtected := app.Group("/auth", middleware.JWTProtected())
	auth.RegisterProctectedRoutes(authProtected, authRepo)

	admin := app.Group("/admin",
		middleware.JWTProtected(),
		middleware.RequireRole("ROLE_ADMIN"),
	)

	auth.RegisterAdminRoutes(admin, authRepo)

	protected := app.Group("/sessions", middleware.JWTProtected())
	sessions.RegisterRoutes(protected, sessionRepo)

	agent := app.Group("/agent", middleware.JWTProtected())
	agentUploadProtection := []fiber.Handler{
		limiter.New(limiter.Config{
			Max: cfg.UploadRatePerMinute,
		}),
		middleware.LimitRequestBody(cfg.UploadMaxPayloadBytes),
	}
	agentapi.RegisterRoutes(agent, sessionRepo, agentUploadProtection...)
	profiles.RegisterRoutes(agent, profilesRepo)

	return app
}
