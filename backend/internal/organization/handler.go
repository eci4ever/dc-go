package organization

import (
	"errors"
	"strings"

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

func (h *Handler) ListOwned(c *fiber.Ctx) error {
	organizations, err := h.svc.ListOwned(c.UserContext(), c.Locals("user_id").(string))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(organizations))
}

func (h *Handler) ListMemberships(c *fiber.Ctx) error {
	organizations, err := h.svc.ListMemberships(
		c.UserContext(), c.Locals("user_id").(string),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(organizations))
}

func (h *Handler) AdminList(c *fiber.Ctx) error {
	orgs, err := h.svc.AdminList(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(orgs))
}

func (h *Handler) AdminCreate(c *fiber.Ctx) error {
	var req CreateOrgRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err.Error()))
	}
	org, err := h.svc.AdminCreate(c.UserContext(), req, c.Locals("user_id").(string))
	if err != nil {
		if errors.Is(err, ErrSlugExists) {
			return c.Status(fiber.StatusConflict).JSON(response.Error("slug already exists"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
	return c.Status(fiber.StatusCreated).JSON(response.Created(org))
}

func (h *Handler) AdminUpdate(c *fiber.Ctx) error {
	var req UpdateOrgRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err.Error()))
	}
	org, err := h.svc.AdminUpdate(c.UserContext(), c.Params("id"), c.Locals("user_id").(string), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.NotFound())
		}
		if errors.Is(err, ErrSlugExists) {
			return c.Status(fiber.StatusConflict).JSON(response.Error("slug already exists"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(org))
}

func (h *Handler) AdminUploadLogo(c *fiber.Ctx) error {
	return h.uploadLogo(c, false)
}

func (h *Handler) UploadLogo(c *fiber.Ctx) error {
	return h.uploadLogo(c, true)
}

func (h *Handler) uploadLogo(c *fiber.Ctx, requireOwner bool) error {
	header, err := c.FormFile("logo")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("logo file is required"))
	}
	file, err := header.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("unable to read logo"))
	}
	defer file.Close()

	var org Organization
	if requireOwner {
		org, err = h.svc.OwnerUploadLogo(
			c.UserContext(), c.Params("id"), c.Locals("user_id").(string), file, header.Size,
		)
	} else {
		org, err = h.svc.AdminUploadLogo(
			c.UserContext(), c.Params("id"), c.Locals("user_id").(string), file, header.Size,
		)
	}
	if err != nil {
		switch {
		case errors.Is(err, ErrLogoTooLarge):
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(response.Error(err.Error()))
		case errors.Is(err, ErrInvalidLogo):
			return c.Status(fiber.StatusBadRequest).JSON(response.Error(err.Error()))
		case errors.Is(err, ErrNotFound):
			return c.Status(fiber.StatusNotFound).JSON(response.NotFound())
		case errors.Is(err, ErrForbidden):
			return c.Status(fiber.StatusForbidden).JSON(response.Error("forbidden"))
		case errors.Is(err, ErrOrganizationLocked):
			return c.Status(fiber.StatusLocked).JSON(response.Error(err.Error()))
		case errors.Is(err, ErrLogoUnavailable):
			return c.Status(fiber.StatusServiceUnavailable).JSON(response.Error("logo storage is unavailable"))
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
		}
	}
	return c.JSON(response.OK(org))
}

func (h *Handler) AdminSetOwner(c *fiber.Ctx) error {
	var req SetOwnerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err.Error()))
	}

	owner, err := h.svc.AdminSetOwner(c.UserContext(), c.Params("id"), c.Locals("user_id").(string), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.NotFound())
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(owner))
}

func (h *Handler) AdminUpdateStatus(c *fiber.Ctx) error {
	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err.Error()))
	}
	org, err := h.svc.AdminUpdateStatus(
		c.UserContext(), c.Params("id"), c.Locals("user_id").(string), req,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.NotFound())
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(org))
}

func (h *Handler) AdminListAudit(c *fiber.Ctx) error {
	logs, err := h.svc.AdminListAudit(c.UserContext(), c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(logs))
}

func (h *Handler) GetLogo(c *fiber.Ctx) error {
	logo, err := h.svc.GetLogo(c.UserContext(), c.Params("id"))
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound), errors.Is(err, ErrNoLogo):
			return c.Status(fiber.StatusNotFound).JSON(response.NotFound())
		case errors.Is(err, ErrLogoUnavailable):
			return c.Status(fiber.StatusServiceUnavailable).JSON(response.Error("logo storage is unavailable"))
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
		}
	}
	if logo.ETag != "" {
		c.Set(fiber.HeaderETag, logo.ETag)
		if strings.TrimSpace(c.Get(fiber.HeaderIfNoneMatch)) == logo.ETag {
			return c.SendStatus(fiber.StatusNotModified)
		}
	}
	c.Set(fiber.HeaderContentType, logo.ContentType)
	c.Set(fiber.HeaderCacheControl, "private, max-age=86400")
	return c.Send(logo.Data)
}

func (h *Handler) AdminDelete(c *fiber.Ctx) error {
	if err := h.svc.AdminDelete(c.UserContext(), c.Params("id")); err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.NotFound())
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
	return c.SendStatus(fiber.StatusNoContent)
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
		switch {
		case errors.Is(err, ErrNotFound):
			return c.Status(404).JSON(response.NotFound())
		case errors.Is(err, ErrForbidden):
			return c.Status(403).JSON(response.Error("forbidden"))
		case errors.Is(err, ErrOrganizationLocked):
			return c.Status(fiber.StatusLocked).JSON(response.Error(err.Error()))
		case errors.Is(err, ErrSlugExists):
			return c.Status(409).JSON(response.Error("slug already exists"))
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
		if errors.Is(err, ErrForbidden) {
			return c.Status(403).JSON(response.Error("forbidden"))
		}
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
		switch {
		case errors.Is(err, ErrForbidden):
			return c.Status(403).JSON(response.Error("forbidden"))
		case errors.Is(err, ErrMemberNotFound):
			return c.Status(404).JSON(response.NotFound())
		case errors.Is(err, ErrOwnerProtected):
			return c.Status(409).JSON(response.Error(err.Error()))
		case errors.Is(err, ErrOrganizationLocked):
			return c.Status(fiber.StatusLocked).JSON(response.Error(err.Error()))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(nil))
}

func (h *Handler) UpdateMemberPermissions(c *fiber.Ctx) error {
	var req UpdateMemberPermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error(err.Error()))
	}
	member, err := h.svc.UpdateMemberPermissions(
		c.UserContext(), c.Params("id"), c.Params("userID"),
		c.Locals("user_id").(string), req.Permissions,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			return c.Status(fiber.StatusForbidden).JSON(response.Error("forbidden"))
		case errors.Is(err, ErrMemberNotFound):
			return c.Status(fiber.StatusNotFound).JSON(response.NotFound())
		case errors.Is(err, ErrOwnerProtected):
			return c.Status(fiber.StatusConflict).JSON(response.Error(err.Error()))
		case errors.Is(err, ErrOrganizationLocked):
			return c.Status(fiber.StatusLocked).JSON(response.Error(err.Error()))
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
		}
	}
	return c.JSON(response.OK(member))
}

func (h *Handler) RemoveMember(c *fiber.Ctx) error {
	orgID := c.Params("id")
	userID := c.Params("userID")
	if err := h.svc.RemoveMember(c.UserContext(), orgID, userID, c.Locals("user_id").(string)); err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			return c.Status(403).JSON(response.Error("forbidden"))
		case errors.Is(err, ErrMemberNotFound):
			return c.Status(404).JSON(response.NotFound())
		case errors.Is(err, ErrOwnerProtected):
			return c.Status(409).JSON(response.Error(err.Error()))
		case errors.Is(err, ErrOrganizationLocked):
			return c.Status(fiber.StatusLocked).JSON(response.Error(err.Error()))
		}
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
		switch {
		case errors.Is(err, ErrForbidden):
			return c.Status(403).JSON(response.Error("forbidden"))
		case errors.Is(err, ErrOrganizationLocked):
			return c.Status(fiber.StatusLocked).JSON(response.Error(err.Error()))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.Status(201).JSON(response.Created(inv))
}

func (h *Handler) ListAudit(c *fiber.Ctx) error {
	logs, err := h.svc.ListAudit(
		c.UserContext(), c.Params("id"), c.Locals("user_id").(string),
	)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			return c.Status(fiber.StatusForbidden).JSON(response.Error("forbidden"))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(logs))
}

func (h *Handler) ListInvitations(c *fiber.Ctx) error {
	orgID := c.Params("id")
	invs, err := h.svc.ListInvitations(c.UserContext(), orgID, c.Locals("user_id").(string))
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			return c.Status(403).JSON(response.Error("forbidden"))
		}
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
		switch {
		case errors.Is(err, ErrForbidden):
			return c.Status(403).JSON(response.Error("forbidden"))
		case errors.Is(err, ErrOrganizationLocked):
			return c.Status(fiber.StatusLocked).JSON(response.Error(err.Error()))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.SendStatus(204)
}
