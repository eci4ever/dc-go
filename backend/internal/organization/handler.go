package organization

import (
	"errors"

	"github.com/eci4ever/dc-go/pkg/response"
	"github.com/eci4ever/dc-go/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateOrgRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	userID := c.Locals("user_id").(string)
	org, err := h.svc.Create(c.UserContext(), req, userID)
	if err != nil {
		if errors.Is(err, ErrSlugExists) {
			return c.Status(409).JSON(response.Error("slug already exists"))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.Status(201).JSON(response.Created(org))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	org, err := h.svc.GetByID(c.UserContext(), id, c.Locals("user_id").(string))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(org))
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	orgs, err := h.svc.List(c.UserContext(), userID)
	if err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(orgs))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateOrgRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	org, err := h.svc.Update(c.UserContext(), id, c.Locals("user_id").(string), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(org))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.Delete(c.UserContext(), id, c.Locals("user_id").(string)); err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.SendStatus(204)
}

func (h *Handler) GetMembers(c *fiber.Ctx) error {
	orgID := c.Params("id")
	members, err := h.svc.GetMembers(c.UserContext(), orgID, c.Locals("user_id").(string))
	if err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(members))
}

func (h *Handler) GetMember(c *fiber.Ctx) error {
	orgID := c.Params("id")
	userID := c.Locals("user_id").(string)
	member, err := h.svc.GetMember(c.UserContext(), orgID, userID)
	if err != nil {
		if errors.Is(err, ErrNotMember) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(member))
}

func (h *Handler) UpdateMemberRole(c *fiber.Ctx) error {
	orgID := c.Params("id")
	userID := c.Params("userID")
	var req UpdateMemberRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	if err := h.svc.UpdateMemberRole(c.UserContext(), orgID, userID, c.Locals("user_id").(string), req.Role); err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(nil))
}

func (h *Handler) RemoveMember(c *fiber.Ctx) error {
	orgID := c.Params("id")
	userID := c.Params("userID")
	if err := h.svc.RemoveMember(c.UserContext(), orgID, userID, c.Locals("user_id").(string)); err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.SendStatus(204)
}

func (h *Handler) Invite(c *fiber.Ctx) error {
	orgID := c.Params("id")
	var req InviteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	userID := c.Locals("user_id").(string)
	inv, err := h.svc.Invite(c.UserContext(), orgID, req.Email, req.Role, userID)
	if err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.Status(201).JSON(response.Created(inv))
}

func (h *Handler) ListInvitations(c *fiber.Ctx) error {
	orgID := c.Params("id")
	invs, err := h.svc.ListInvitations(c.UserContext(), orgID, c.Locals("user_id").(string))
	if err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(invs))
}

func (h *Handler) AcceptInvitation(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(string)
	if err := h.svc.AcceptInvitation(c.UserContext(), id, userID); err != nil {
		if errors.Is(err, ErrInvitationExpired) || errors.Is(err, ErrInvitationNotFound) {
			return c.Status(400).JSON(response.Error(err.Error()))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(nil))
}

func (h *Handler) DeclineInvitation(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.DeclineInvitation(c.UserContext(), id, c.Locals("user_id").(string)); err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(nil))
}

func (h *Handler) CancelInvitation(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.CancelInvitation(c.UserContext(), id, c.Locals("user_id").(string)); err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.SendStatus(204)
}
