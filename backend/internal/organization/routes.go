package organization

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, h *Handler, mw ...fiber.Handler) {
	if len(mw) > 0 && mw[0] != nil {
		router.Post("/organizations", mw[0], mw[1], h.Create)
		router.Get("/organizations", mw[0], h.List)
		router.Get("/organizations/:id", mw[0], h.GetByID)
		router.Put("/organizations/:id", mw[0], mw[1], h.Update)
		router.Delete("/organizations/:id", mw[0], mw[1], h.Delete)
		router.Get("/organizations/:id/members", mw[0], h.GetMembers)
		router.Get("/organizations/:id/members/me", mw[0], h.GetMember)
		router.Put("/organizations/:id/members/:userID/role", mw[0], mw[1], h.UpdateMemberRole)
		router.Delete("/organizations/:id/members/:userID", mw[0], mw[1], h.RemoveMember)
		router.Post("/organizations/:id/invitations", mw[0], mw[1], h.Invite)
		router.Get("/organizations/:id/invitations", mw[0], h.ListInvitations)
		router.Post("/invitations/:id/accept", mw[0], mw[1], h.AcceptInvitation)
		router.Post("/invitations/:id/decline", mw[0], mw[1], h.DeclineInvitation)
		router.Delete("/invitations/:id", mw[0], mw[1], h.CancelInvitation)
		return
	}
	router.Post("/organizations", h.Create)
	router.Get("/organizations", h.List)
	router.Get("/organizations/:id", h.GetByID)
	router.Put("/organizations/:id", h.Update)
	router.Delete("/organizations/:id", h.Delete)
	router.Get("/organizations/:id/members", h.GetMembers)
	router.Get("/organizations/:id/members/me", h.GetMember)
	router.Put("/organizations/:id/members/:userID/role", h.UpdateMemberRole)
	router.Delete("/organizations/:id/members/:userID", h.RemoveMember)
	router.Post("/organizations/:id/invitations", h.Invite)
	router.Get("/organizations/:id/invitations", h.ListInvitations)
	router.Post("/invitations/:id/accept", h.AcceptInvitation)
	router.Post("/invitations/:id/decline", h.DeclineInvitation)
	router.Delete("/invitations/:id", h.CancelInvitation)
}
