package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/domain/parts"
	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/internal/infrastructure/middleware"
	"github.com/nova/backend/pkg/errors"
)

type PartHandler struct {
	service *parts.PartService
}

func NewPartHandler(service *parts.PartService) *PartHandler {
	return &PartHandler{service: service}
}

func (h *PartHandler) List(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	org := c.Query("org")
	pagination := dto.PaginationQuery{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 20),
	}

	parts, total, err := h.service.FindAll(c.Context(), tenant, org, pagination.GetLimit(), pagination.GetOffset())
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewListResponse(parts, dto.NewPaginationMeta(pagination.Page, pagination.PageSize, total)))
}

func (h *PartHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	part, err := h.service.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	if part == nil {
		return c.Status(404).JSON(dto.NewErrorResponse("NOT_FOUND", "Part not found"))
	}

	return c.JSON(dto.NewSuccessResponse(part))
}

func (h *PartHandler) Create(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	var req parts.CreatePartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	part, err := h.service.Create(c.Context(), tenant, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(part))
}

func (h *PartHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req parts.UpdatePartRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	part, err := h.service.Update(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(part))
}

func (h *PartHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("Part deleted successfully"))
}
