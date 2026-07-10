package team

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, mw ...fiber.Handler) {
	if len(mw) > 0 && mw[0] != nil {
		router.Post("/organizations/:orgID/teams", mw[0], h.Create)
		router.Get("/organizations/:orgID/teams", mw[0], h.List)
		router.Get("/teams/:id", mw[0], h.GetByID)
		router.Put("/teams/:id", mw[0], h.Update)
		router.Delete("/teams/:id", mw[0], h.Delete)
		router.Post("/teams/:id/members", mw[0], h.AddMember)
		router.Get("/teams/:id/members", mw[0], h.GetMembers)
		router.Delete("/teams/:id/members/:userID", mw[0], h.RemoveMember)
		return
	}
	router.Post("/organizations/:orgID/teams", h.Create)
	router.Get("/organizations/:orgID/teams", h.List)
	router.Get("/teams/:id", h.GetByID)
	router.Put("/teams/:id", h.Update)
	router.Delete("/teams/:id", h.Delete)
	router.Post("/teams/:id/members", h.AddMember)
	router.Get("/teams/:id/members", h.GetMembers)
	router.Delete("/teams/:id/members/:userID", h.RemoveMember)
}
