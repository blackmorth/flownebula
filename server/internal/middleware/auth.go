package middleware

import (
	"flownebula/server/internal/metrics"
	"strings"

	"flownebula/server/internal/auth"

	"github.com/gofiber/fiber/v2"
)

func JWTProtected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		metrics.ServerMetrics.HTTPRequestsTotal.Add(1)
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			metrics.ServerMetrics.HTTPUnauthorizedTotal.Add(1)
			addCORSHeaders(c)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing Authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			metrics.ServerMetrics.HTTPUnauthorizedTotal.Add(1)
			addCORSHeaders(c)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid Authorization header",
			})
		}

		tokenStr := parts[1]

		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			metrics.ServerMetrics.HTTPUnauthorizedTotal.Add(1)
			addCORSHeaders(c)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or expired token",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("roles", claims.Roles)

		return c.Next()
	}
}

func addCORSHeaders(c *fiber.Ctx) {
	origin := c.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	c.Set("Access-Control-Allow-Origin", origin)
	c.Set("Access-Control-Allow-Credentials", "true")
}

func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roles := c.Locals("roles").([]string)

		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}
		addCORSHeaders(c)
		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}
}
