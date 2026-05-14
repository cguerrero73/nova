package structure

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/internal/infrastructure/middleware"
	"github.com/nova/backend/pkg/errors"
)

type StructureHandler struct {
	service *StructureService
}

func NewStructureHandler(service *StructureService) *StructureHandler {
	return &StructureHandler{service: service}
}

type CreateStructureRequest struct {
	ParentCode string `json:"sct_parent_code" validate:"required"`
	ParentOrg  string `json:"sct_parent_org" validate:"required"`
	ChildCode  string `json:"sct_child_code" validate:"required"`
	ChildOrg   string `json:"sct_child_org" validate:"required"`
	Cost       string `json:"sct_cost"`
	Meter      string `json:"sct_meter"`
}

type UpdateStructureRequest struct {
	Cost  string `json:"sct_cost"`
	Meter string `json:"sct_meter"`
}

func (h *StructureHandler) List(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	structures, err := h.service.FindAll(c.Context(), tenant)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(structures))
}

func (h *StructureHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	structure, err := h.service.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	if structure == nil {
		return c.Status(404).JSON(dto.NewErrorResponse("NOT_FOUND", "Structure not found"))
	}

	return c.JSON(dto.NewSuccessResponse(structure))
}

func (h *StructureHandler) Create(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	var req CreateStructureRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	structure, err := h.service.Create(c.Context(), tenant, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(structure))
}

func (h *StructureHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateStructureRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	structure, err := h.service.Update(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(structure))
}

func (h *StructureHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("Structure deleted successfully"))
}
