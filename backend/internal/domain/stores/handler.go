package stores

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/internal/infrastructure/middleware"
	"github.com/nova/backend/pkg/errors"
)

type StoreHandler struct {
	service *StoreService
}

func NewStoreHandler(service *StoreService) *StoreHandler {
	return &StoreHandler{service: service}
}

type CreateStoreRequest struct {
	Code string `json:"str_code" validate:"required"`
	Name string `json:"str_name" validate:"required"`
	Desc string `json:"str_desc"`
	Org  string `json:"str_org" validate:"required"`
}

type UpdateStoreRequest struct {
	Name string `json:"str_name"`
	Desc string `json:"str_desc"`
}

type CreateBinRequest struct {
	Code string `json:"bin_code" validate:"required"`
	Desc string `json:"bin_desc"`
	Org  string `json:"bin_org" validate:"required"`
}

type UpdateBinRequest struct {
	Desc string `json:"bin_desc"`
}

func (h *StoreHandler) List(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	org := c.Query("org")
	stores, err := h.service.FindAll(c.Context(), tenant, org)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(stores))
}

func (h *StoreHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	store, err := h.service.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	if store == nil {
		return c.Status(404).JSON(dto.NewErrorResponse("NOT_FOUND", "Store not found"))
	}

	return c.JSON(dto.NewSuccessResponse(store))
}

func (h *StoreHandler) Create(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	var req CreateStoreRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	store, err := h.service.Create(c.Context(), tenant, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(store))
}

func (h *StoreHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateStoreRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	store, err := h.service.Update(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(store))
}

func (h *StoreHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("Store deleted successfully"))
}

// Bin handlers
func (h *StoreHandler) ListBins(c *fiber.Ctx) error {
	org := c.Query("org")
	bins, err := h.service.FindBinsByOrg(c.Context(), org)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(bins))
}

func (h *StoreHandler) CreateBin(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	var req CreateBinRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	bin, err := h.service.CreateBin(c.Context(), tenant, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(bin))
}

func (h *StoreHandler) UpdateBin(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateBinRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	bin, err := h.service.UpdateBin(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(bin))
}

func (h *StoreHandler) DeleteBin(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.DeleteBin(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("Bin deleted successfully"))
}
