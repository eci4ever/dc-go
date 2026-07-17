package auth

import "github.com/gofiber/fiber/v2"

type RateLimiters struct {
	Register       fiber.Handler
	Login          fiber.Handler
	Refresh        fiber.Handler
	ChangePassword fiber.Handler
}

func RegisterRoutes(router fiber.Router, h *Handler, authMw, csrfMw fiber.Handler, limits RateLimiters) {
	router.Post("/auth/register", limits.Register, h.Register)
	router.Post("/auth/login", limits.Login, h.Login)
	router.Post("/auth/refresh", csrfMw, limits.Refresh, h.Refresh)
	router.Get("/auth/session", authMw, h.GetSession)
	router.Put("/auth/session/active-organization", authMw, csrfMw, h.SetActiveOrganization)
	router.Put("/auth/password", authMw, csrfMw, limits.ChangePassword, h.ChangePassword)
	router.Get("/auth/sessions", authMw, h.ListSessions)
	router.Delete("/auth/sessions/:id", authMw, csrfMw, h.RevokeSession)
	router.Post("/auth/logout", csrfMw, h.Logout)
}
