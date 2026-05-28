package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/domain/auth"
	"github.com/nova/backend/pkg/errors"
)

type AuthHandler struct {
	authService *auth.AuthService
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	tenant := c.Locals("tenant").(string)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired)
	}

	var req auth.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	req.Tenant = tenant

	resp, err := h.authService.Login(c.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(errors.ErrInternal)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	tenant := c.Locals("tenant").(string)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired)
	}

	var req auth.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	req.Tenant = tenant

	resp, err := h.authService.Register(c.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(errors.ErrInternal)
	}

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req auth.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	resp, err := h.authService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(errors.ErrInternal)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	user := c.Locals("user").(*auth.TokenClaims)

	if err := h.authService.Logout(c.Context(), user.UserCode); err != nil {
		return c.Status(500).JSON(errors.ErrInternal)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"message": "Logged out successfully",
		},
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user := c.Locals("user").(*auth.TokenClaims)

	// Get full user from service
	resp, err := h.authService.GetUserByCode(c.Context(), user.UserCode)
	if err != nil {
		return c.Status(500).JSON(errors.ErrInternal)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}
