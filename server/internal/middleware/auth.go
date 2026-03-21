package middleware

import (
	"strings"

	"flownebula/server/internal/auth"

	"github.com/gofiber/fiber/v2"
)

func JWTProtected() fiber.Handler {
    return func(c *fiber.Ctx) error {
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "missing Authorization header",
            })
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "invalid Authorization header",
            })
        }

        tokenStr := parts[1]

        claims, err := auth.ParseToken(tokenStr)
        if err != nil {
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


func RequireRole(role string) fiber.Handler {
    return func(c *fiber.Ctx) error {
        roles := c.Locals("roles").([]string)

        for _, r := range roles {
            if r == role {
                return c.Next()
            }
        }

        return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
    }
}
