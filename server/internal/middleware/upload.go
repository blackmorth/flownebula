package middleware

import "github.com/gofiber/fiber/v2"

func LimitRequestBody(maxBytes int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if maxBytes > 0 && len(c.Body()) > maxBytes {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": "payload too large",
			})
		}

		return c.Next()
	}
}
