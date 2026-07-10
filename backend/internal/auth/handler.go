package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"dc-express/pkg/response"
	"dc-express/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	svc    *Service
	secure bool
}

func NewHandler(svc *Service, secure bool) *Handler {
	return &Handler{svc: svc, secure: secure}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	tokens, err := h.svc.Register(c.UserContext(), req)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			return c.Status(409).JSON(response.Error("email already exists"))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	h.setCookies(c, tokens)
	return c.Status(201).JSON(response.Created(SessionResponse{User: tokens.User, Session: tokens.Session}))
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(response.Error("invalid request body"))
	}
	if err := validator.Validate(req); err != nil {
		return c.Status(400).JSON(response.Error(err.Error()))
	}

	tokens, err := h.svc.Login(c.UserContext(), req, c.IP(), c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrUserBanned) {
			return c.Status(401).JSON(response.Error(err.Error()))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	h.setCookies(c, tokens)
	return c.JSON(response.OK(SessionResponse{User: tokens.User, Session: tokens.Session}))
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	tokens, err := h.svc.Refresh(c.UserContext(), c.Cookies("refresh_token"))
	if err != nil {
		if errors.Is(err, ErrInvalidToken) || errors.Is(err, ErrExpiredToken) {
			return c.Status(401).JSON(response.Error(err.Error()))
		}
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	h.setCookies(c, tokens)
	return c.JSON(response.OK(SessionResponse{User: tokens.User, Session: tokens.Session}))
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
	if err := h.svc.Logout(c.UserContext(), c.Cookies("refresh_token")); err != nil {
		return c.Status(500).JSON(response.Error("internal server error"))
	}

	h.clearCookies(c)
	return c.JSON(response.OK(nil))
}

func (h *Handler) setCookies(c *fiber.Ctx, tokens TokenResponse) {
	csrf := make([]byte, 32)
	_, _ = rand.Read(csrf)
	c.Cookie(&fiber.Cookie{Name: "access_token", Value: tokens.AccessToken, HTTPOnly: true, Secure: h.secure, SameSite: "Strict", Path: "/", MaxAge: int(accessTokenTTL.Seconds())})
	c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: tokens.RefreshToken, HTTPOnly: true, Secure: h.secure, SameSite: "Strict", Path: "/api/v1/auth", MaxAge: int(refreshTokenTTL.Seconds())})
	c.Cookie(&fiber.Cookie{Name: "csrf_token", Value: base64.RawURLEncoding.EncodeToString(csrf), HTTPOnly: false, Secure: h.secure, SameSite: "Strict", Path: "/", MaxAge: int(refreshTokenTTL.Seconds())})
}
func (h *Handler) clearCookies(c *fiber.Ctx) {
	for _, name := range []string{"access_token", "csrf_token"} {
		c.Cookie(&fiber.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1, HTTPOnly: name == "access_token", Secure: h.secure, SameSite: "Strict"})
	}
	c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: "", Path: "/api/v1/auth", MaxAge: -1, HTTPOnly: true, Secure: h.secure, SameSite: "Strict"})
}
