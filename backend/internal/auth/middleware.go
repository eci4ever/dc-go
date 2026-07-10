package auth

import (
	"strings"

	"dc-express/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwt *JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).JSON(response.Error("missing authorization header"))
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).JSON(response.Error("invalid authorization format"))
		}

		claims, err := jwt.Verify(parts[1])
		if err != nil {
			return c.Status(401).JSON(response.Error("invalid or expired token"))
		}

		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}
