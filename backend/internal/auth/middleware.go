package auth

import (
	"context"
	"crypto/subtle"

	"github.com/eci4ever/dc-go/internal/user"
	"github.com/eci4ever/dc-go/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type UserRoleLookup interface {
	GetByID(context.Context, string) (user.User, error)
}

func AuthMiddleware(jwt *JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := jwt.Verify(c.Cookies("access_token"))
		if err != nil {
			return c.Status(401).JSON(response.Error("invalid or expired token"))
		}

		c.Locals("user_id", claims.Subject)
		return c.Next()
	}
}

func RequireUserRole(repo UserRoleLookup, role user.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, _ := c.Locals("user_id").(string)
		u, err := repo.GetByID(c.UserContext(), userID)
		if err != nil {
			return c.Status(500).JSON(response.Error("internal server error"))
		}
		if u.Role != role {
			return c.Status(403).JSON(response.Error("forbidden"))
		}
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
