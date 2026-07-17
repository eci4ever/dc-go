package team

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
	orgID := c.Params("orgID")
	var req CreateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	team, err := h.svc.Create(c.UserContext(), orgID, c.Locals("user_id").(string), req)
	if err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.Status(201).JSON(response.Created(team))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	team, err := h.svc.GetByID(c.UserContext(), id, c.Locals("user_id").(string))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(team))
}

func (h *Handler) List(c *fiber.Ctx) error {
	orgID := c.Params("orgID")
	teams, err := h.svc.List(c.UserContext(), orgID, c.Locals("user_id").(string))
	if err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(teams))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateTeamRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	team, err := h.svc.Update(c.UserContext(), id, c.Locals("user_id").(string), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(team))
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

func (h *Handler) AddMember(c *fiber.Ctx) error {
	teamID := c.Params("id")
	var req AddMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	member, err := h.svc.AddMember(c.UserContext(), teamID, c.Locals("user_id").(string), req)
	if err != nil {
		if errors.Is(err, ErrAlreadyMember) {
			return c.Status(409).JSON(response.Error("already a member"))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.Status(201).JSON(response.Created(member))
}

func (h *Handler) GetMembers(c *fiber.Ctx) error {
	teamID := c.Params("id")
	members, err := h.svc.GetMembers(c.UserContext(), teamID, c.Locals("user_id").(string))
	if err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.JSON(response.OK(members))
}

func (h *Handler) RemoveMember(c *fiber.Ctx) error {
	teamID := c.Params("id")
	userID := c.Params("userID")
	if err := h.svc.RemoveMember(c.UserContext(), teamID, userID, c.Locals("user_id").(string)); err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}
	return c.SendStatus(204)
}
