package auth

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, authMw fiber.Handler) {
	router.Post("/auth/register", h.Register)
	router.Post("/auth/login", h.Login)
	router.Post("/auth/refresh", h.Refresh)
	router.Get("/auth/session", authMw, h.GetSession)
	router.Post("/auth/logout", authMw, h.Logout)
}
