package user

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

func (h *Handler) UploadAvatar(c *fiber.Ctx) error {
	id, _ := c.Locals("user_id").(string)
	if id == "" {
		return c.Status(400).JSON(response.Error("invalid id"))
	}
	header, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(400).JSON(response.Error("avatar file is required"))
	}
	file, err := header.Open()
	if err != nil {
		return c.Status(400).JSON(response.Error("unable to read avatar"))
	}
	defer file.Close()

	u, err := h.svc.UploadAvatar(c.UserContext(), id, file, header.Size)
	if err != nil {
		switch {
		case errors.Is(err, ErrAvatarTooLarge):
			return c.Status(413).JSON(response.Error(err.Error()))
		case errors.Is(err, ErrInvalidAvatar):
			return c.Status(400).JSON(response.Error(err.Error()))
		case errors.Is(err, ErrNotFound):
			return c.Status(404).JSON(response.NotFound())
		case errors.Is(err, ErrAvatarUnavailable):
			return c.Status(503).JSON(response.Error("avatar storage is unavailable"))
		default:
			return c.Status(500).JSON(response.Error("internal server error"))
		}
	}
	return c.JSON(response.OK(u))
}

func (h *Handler) GetAvatar(c *fiber.Ctx) error {
	avatar, err := h.svc.GetAvatar(c.UserContext(), c.Params("id"))
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound), errors.Is(err, ErrNoAvatar):
			return c.Status(404).JSON(response.NotFound())
		case errors.Is(err, ErrAvatarUnavailable):
			return c.Status(503).JSON(response.Error("avatar storage is unavailable"))
		default:
			return c.Status(500).JSON(response.Error("internal server error"))
		}
	}
	if avatar.ETag != "" {
		c.Set(fiber.HeaderETag, avatar.ETag)
		if strings.TrimSpace(c.Get(fiber.HeaderIfNoneMatch)) == avatar.ETag {
			return c.SendStatus(fiber.StatusNotModified)
		}
	}
	c.Set(fiber.HeaderContentType, avatar.ContentType)
	c.Set(fiber.HeaderCacheControl, "private, max-age=86400")
	return c.Send(avatar.Data)
}

func (h *Handler) RemoveAvatar(c *fiber.Ctx) error {
	id, _ := c.Locals("user_id").(string)
	if id == "" {
		return c.Status(400).JSON(response.Error("invalid id"))
	}
	u, err := h.svc.RemoveAvatar(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(u))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, _ := c.Locals("user_id").(string)
	if id == "" {
		return c.Status(400).JSON(response.Error("invalid id"))
	}

	u, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.JSON(response.OK(u))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, _ := c.Locals("user_id").(string)
	if id == "" {
		return c.Status(400).JSON(response.Error("invalid id"))
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	u, err := h.svc.Update(c.UserContext(), id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.JSON(response.OK(u))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, _ := c.Locals("user_id").(string)
	if id == "" {
		return c.Status(400).JSON(response.Error("invalid id"))
	}

	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.SendStatus(204)
}

func (h *Handler) List(c *fiber.Ctx) error {
	users, err := h.svc.List(c.UserContext())
	if err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(users))
}

func (h *Handler) UpdateRole(c *fiber.Ctx) error {
	var req UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	u, err := h.svc.UpdateRole(c.UserContext(), c.Params("id"), c.Locals("user_id").(string), req.Role)
	if err != nil {
		switch {
		case errors.Is(err, ErrSelfRole):
			return c.Status(400).JSON(response.Error(err.Error()))
		case errors.Is(err, ErrNotFound):
			return c.Status(404).JSON(response.NotFound())
		default:
			return c.Status(500).JSON(response.Error("internal server error"))
		}
	}
	return c.JSON(response.OK(u))
}
