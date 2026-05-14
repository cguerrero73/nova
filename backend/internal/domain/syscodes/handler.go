package syscodes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/pkg/errors"
)

type SysCodeHandler struct {
	service *SysCodeService
}

func NewSysCodeHandler(service *SysCodeService) *SysCodeHandler {
	return &SysCodeHandler{service: service}
}

type CreateSysCodeRequest struct {
	Type   string `json:"sys_type" validate:"required"`
	Code   string `json:"sys_code" validate:"required"`
	UCode  string `json:"sys_ucode" validate:"required"`
	Desc   string `json:"sys_desc"`
	System string `json:"sys_system"`
}

type UpdateSysCodeRequest struct {
	UCode string `json:"sys_ucode"`
	Desc  string `json:"sys_desc"`
}

func (h *SysCodeHandler) List(c *fiber.Ctx) error {
	syscodes, err := h.service.FindAll(c.Context())
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(syscodes))
}

func (h *SysCodeHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	syscode, err := h.service.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	if syscode == nil {
		return c.Status(404).JSON(dto.NewErrorResponse("NOT_FOUND", "SysCode not found"))
	}

	return c.JSON(dto.NewSuccessResponse(syscode))
}

func (h *SysCodeHandler) Create(c *fiber.Ctx) error {
	var req CreateSysCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	syscode, err := h.service.Create(c.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(syscode))
}

func (h *SysCodeHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateSysCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	syscode, err := h.service.Update(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(syscode))
}

func (h *SysCodeHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(c.Context(), id); err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("SysCode deleted successfully"))
}

func (h *SysCodeHandler) GetByType(c *fiber.Ctx) error {
	codeType := c.Params("type")

	syscodes, err := h.service.FindByType(c.Context(), codeType)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(syscodes))
}
