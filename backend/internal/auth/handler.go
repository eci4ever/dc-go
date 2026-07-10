package auth

import (
	"errors"

	"dc-express/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}

	tokens, err := h.svc.Register(c.UserContext(), req)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			return c.Status(409).JSON(response.Error("email already exists"))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.Status(201).JSON(response.Created(tokens))
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}

	tokens, err := h.svc.Login(c.UserContext(), req, c.IP(), c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrUserBanned) {
			return c.Status(401).JSON(response.Error(err.Error()))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.JSON(response.OK(tokens))
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}

	tokens, err := h.svc.Refresh(c.UserContext(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrExpiredToken) {
			return c.Status(401).JSON(response.Error(err.Error()))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.JSON(response.OK(tokens))
}

func (h *Handler) GetSession(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	sess, err := h.svc.GetSession(c.UserContext(), userID)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			return c.Status(401).JSON(response.Error("unauthorized"))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.JSON(response.OK(sess))
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}

	if err := h.svc.Logout(c.UserContext(), req.RefreshToken); err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	return c.JSON(response.OK(nil))
}
