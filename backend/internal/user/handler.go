package user

import (
	"errors"
	"strconv"

	"dc-express/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}

	u, err := h.svc.Create(c.UserContext(), req)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			return c.Status(409).JSON(response.Error("email already exists"))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.Status(201).JSON(response.Created(u))
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(response.Error("invalid id"))
	}

	u, err := h.svc.GetByID(c.UserContext(), int32(id))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.JSON(response.OK(u))
}

func (h *Handler) List(c *fiber.Ctx) error {
	users, err := h.svc.List(c.UserContext())
	if err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.JSON(response.OK(users))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(response.Error("invalid id"))
	}

	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}

	u, err := h.svc.Update(c.UserContext(), int32(id), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.JSON(response.OK(u))
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(response.Error("invalid id"))
	}

	if err := h.svc.Delete(c.UserContext(), int32(id)); err != nil {
		if errors.Is(err, ErrNotFound) {
			return c.Status(404).JSON(response.NotFound())
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.SendStatus(204)
}
