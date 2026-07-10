package user

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler) {
	router.Get("/users", h.List)
	router.Get("/users/:id", h.GetByID)
	router.Post("/users", h.Create)
	router.Put("/users/:id", h.Update)
	router.Delete("/users/:id", h.Delete)
}
