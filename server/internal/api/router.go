package api

import (
	"flownebula/server/internal/agentapi"
	"flownebula/server/internal/auth"
	"flownebula/server/internal/db"
	"flownebula/server/internal/metrics"
	"flownebula/server/internal/middleware"
	"flownebula/server/internal/sessions"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func New() *fiber.App {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type, Authorization",
	}))

	sqlite := db.Open("nebula.db")
	db.Migrate(sqlite)

	authRepo := auth.NewSQLiteRepo(sqlite)
	sessionRepo := sessions.NewSQLiteRepo(sqlite)
	metricsRepo := metrics.NewSQLiteRepo(sqlite)

	auth.RegisterRoutes(app, authRepo)

	protected := app.Group("/sessions", middleware.JWTProtected())
	sessions.RegisterRoutes(protected, sessionRepo)

	agent := app.Group("/agent", middleware.JWTProtected())
	agentapi.RegisterRoutes(agent, sessionRepo)

	metricsGroup := agent.Group("/metrics")
	metrics.RegisterRoutes(metricsGroup, metricsRepo)

    admin := app.Group("/admin",
        middleware.JWTProtected(),
        middleware.RequireRole("ROLE_ADMIN"),
    )

    auth.RegisterAdminRoutes(admin, authRepo)

	return app
}

