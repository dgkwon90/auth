package middleware

import (
	"auth/internal/service"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JwtMiddleware(jwtSvc *service.JwtService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing token"})
		}
		// "Bearer <token>"
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token format"})
		}
		userId, err := jwtSvc.ValidateAccessToken(parts[1])
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": err.Error()})
		}
		c.Locals("userId", userId)
		return c.Next()
	}
}
