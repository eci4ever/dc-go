package auth

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/eci4ever/dc-go/internal/user"

	"github.com/gofiber/fiber/v2"
)

type fakeUserRoleLookup struct {
	role user.Role
	err  error
}

func (f fakeUserRoleLookup) GetByID(context.Context, string) (user.User, error) {
	return user.User{Role: f.role}, f.err
}

func TestRequireUserRole(t *testing.T) {
	tests := []struct {
		name   string
		lookup fakeUserRoleLookup
		status int
	}{
		{name: "admin accepted", lookup: fakeUserRoleLookup{role: user.RoleAdmin}, status: fiber.StatusNoContent},
		{name: "user forbidden", lookup: fakeUserRoleLookup{role: user.RoleUser}, status: fiber.StatusForbidden},
		{name: "lookup failure", lookup: fakeUserRoleLookup{err: errors.New("database unavailable")}, status: fiber.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				c.Locals("user_id", "user-id")
				return c.Next()
			}, RequireUserRole(tt.lookup, user.RoleAdmin), func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusNoContent)
			})

			response, err := app.Test(httptest.NewRequest("GET", "/", nil))
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != tt.status {
				t.Fatalf("status = %d, want %d", response.StatusCode, tt.status)
			}
		})
	}
}
