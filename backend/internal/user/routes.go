package user

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, authMw, csrfMw, adminMw fiber.Handler) {
	g := router.Group("/users")
	g.Use(authMw)
	g.Get("/me", h.GetByID)
	g.Put("/me", csrfMw, h.Update)
	g.Put("/me/avatar", csrfMw, h.UploadAvatar)
	g.Delete("/me/avatar", csrfMw, h.RemoveAvatar)
	g.Get("/:id/avatar", h.GetAvatar)
	g.Delete("/me", csrfMw, h.Delete)
	g.Get("", adminMw, h.List)
	g.Put("/:id/role", csrfMw, adminMw, h.UpdateRole)
}
