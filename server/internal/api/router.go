package api

import (
	"flownebula/server/internal/auth"
	"flownebula/server/internal/db"
	"flownebula/server/internal/middleware"
	"flownebula/server/internal/scripts"
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

	auth.RegisterRoutes(app, authRepo)
	authProtected := app.Group("/auth", middleware.JWTProtected())
	auth.RegisterProctectedRoutes(authProtected, authRepo)

	protectedSessions := app.Group("/sessions", middleware.JWTProtected())
	sessions.RegisterRoutes(protectedSessions, sessionRepo)

	protectedScripts := app.Group("/scripts", middleware.JWTProtected())
	scripts.RegisterRoutes(protectedScripts, sessionRepo)

	return app
}
