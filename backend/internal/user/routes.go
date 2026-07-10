package user

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, mw ...fiber.Handler) {
	g := router.Group("/users")
	if len(mw) > 0 && mw[0] != nil {
		g.Use(mw[0])
	}
	g.Get("/me", h.GetByID)
	if len(mw) > 1 {
		g.Put("/me", mw[1], h.Update)
		g.Delete("/me", mw[1], h.Delete)
	} else {
		g.Put("/me", h.Update)
		g.Delete("/me", h.Delete)
	}
}
