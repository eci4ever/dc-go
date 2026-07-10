package auth

import (
	"crypto/subtle"

	"dc-express/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwt *JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := jwt.Verify(c.Cookies("access_token"))
		if err != nil {
			return c.Status(401).JSON(response.Error("invalid or expired token"))
		}

		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}

func CSRFMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == fiber.MethodGet || c.Method() == fiber.MethodHead || c.Method() == fiber.MethodOptions {
			return c.Next()
		}
		cookie, header := c.Cookies("csrf_token"), c.Get("X-CSRF-Token")
		if cookie == "" || header == "" || subtle.ConstantTimeCompare([]byte(cookie), []byte(header)) != 1 {
			return c.Status(403).JSON(response.Error("invalid CSRF token"))
		}
		return c.Next()
	}
}
