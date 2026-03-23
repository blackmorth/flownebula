package api

import (
	"flownebula/server/internal/agentapi"
	"flownebula/server/internal/auth"
	"flownebula/server/internal/db"
	"flownebula/server/internal/metrics"
	"flownebula/server/internal/middleware"
	"flownebula/server/internal/profiles"
	"flownebula/server/internal/sessions"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func New() *fiber.App {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:8081",
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

	authRepo := auth.NewSQLiteRepo(sqlite)
	sessionRepo := sessions.NewSQLiteRepo(sqlite)
	metricsRepo := metrics.NewSQLiteRepo(sqlite)
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
	agentapi.RegisterRoutes(agent, sessionRepo)
	profiles.RegisterRoutes(agent, profilesRepo)

	metricsGroup := agent.Group("/metrics")
	metrics.RegisterRoutes(metricsGroup, metricsRepo)

	return app
}
