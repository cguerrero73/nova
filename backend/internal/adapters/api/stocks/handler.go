package api

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/domain/stocks"
	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/internal/infrastructure/middleware"
	"github.com/nova/backend/pkg/errors"
)

type StockHandler struct {
	service *stocks.StockService
}

func NewStockHandler(service *stocks.StockService) *StockHandler {
	return &StockHandler{service: service}
}

func (h *StockHandler) List(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	storeCode := c.Query("store_code")
	storeOrg := c.Query("store_org")

	stocks, err := h.service.FindAll(c.Context(), tenant, storeCode, storeOrg)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(stocks))
}

func (h *StockHandler) GetLowStock(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	stocks, err := h.service.FindLowStock(c.Context(), tenant)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(stocks))
}

func (h *StockHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	stock, err := h.service.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	if stock == nil {
		return c.Status(404).JSON(dto.NewErrorResponse("NOT_FOUND", "Stock not found"))
	}

	return c.JSON(dto.NewSuccessResponse(stock))
}

func (h *StockHandler) Create(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	var req stocks.CreateStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	stock, err := h.service.Create(c.Context(), tenant, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(stock))
}

func (h *StockHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req stocks.UpdateStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	stock, err := h.service.Update(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(stock))
}

func (h *StockHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("Stock deleted successfully"))
}

func (h *StockHandler) AdjustQuantity(c *fiber.Ctx) error {
	id := c.Params("id")

	var req struct {
		Quantity   float64 `json:"quantity" validate:"required"`
		Adjustment float64 `json:"adjustment"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	stock, err := h.service.AdjustQuantity(c.Context(), id, req.Quantity, req.Adjustment)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(stock))
}

// BinStock handlers
func (h *StockHandler) ListBinStocks(c *fiber.Ctx) error {
	partCode := c.Query("part_code")
	partOrg := c.Query("part_org")
	storeCode := c.Query("store_code")
	storeOrg := c.Query("store_org")

	if partCode == "" || storeCode == "" {
		return c.Status(400).JSON(dto.NewErrorResponse("BAD_REQUEST", "part_code and store_code are required"))
	}

	binStocks, err := h.service.FindBinStocksByPartAndStore(c.Context(), partCode, partOrg, storeCode, storeOrg)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(binStocks))
}

func (h *StockHandler) CreateBinStock(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	var req stocks.CreateBinStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	binStock, err := h.service.CreateBinStock(c.Context(), tenant, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(binStock))
}

func (h *StockHandler) UpdateBinStock(c *fiber.Ctx) error {
	id := c.Params("id")

	var req stocks.UpdateBinStockRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	binStock, err := h.service.UpdateBinStock(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(binStock))
}

func (h *StockHandler) DeleteBinStock(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.DeleteBinStock(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("BinStock deleted successfully"))
}
