package user

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, mw ...fiber.Handler) {
	g := router.Group("/users")
	if len(mw) > 0 && mw[0] != nil {
		g.Use(mw[0])
	}
	g.Get("/", h.List)
	g.Get("/:id", h.GetByID)
	g.Put("/:id", h.Update)
	g.Delete("/:id", h.Delete)
}
