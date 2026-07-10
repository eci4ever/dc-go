package auth

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, authMw, csrfMw fiber.Handler) {
	router.Post("/auth/register", h.Register)
	router.Post("/auth/login", h.Login)
	router.Post("/auth/refresh", csrfMw, h.Refresh)
	router.Get("/auth/session", authMw, h.GetSession)
	router.Put("/auth/session/active-organization", authMw, csrfMw, h.SetActiveOrganization)
	router.Put("/auth/password", authMw, csrfMw, h.ChangePassword)
	router.Get("/auth/sessions", authMw, h.ListSessions)
	router.Delete("/auth/sessions/:id", authMw, csrfMw, h.RevokeSession)
	router.Post("/auth/logout", csrfMw, h.Logout)
}
