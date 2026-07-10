package auth

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, authMw, csrfMw fiber.Handler, secure bool) {
	router.Post("/auth/register", h.Register)
	router.Post("/auth/login", h.Login)
	router.Post("/auth/refresh", csrfMw, h.Refresh)
	router.Get("/auth/session", authMw, h.GetSession)
	router.Post("/auth/logout", authMw, csrfMw, h.Logout)
}
