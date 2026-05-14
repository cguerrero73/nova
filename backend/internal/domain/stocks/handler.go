package stocks

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/internal/infrastructure/middleware"
	"github.com/nova/backend/pkg/errors"
)

type StockHandler struct {
	service *StockService
}

func NewStockHandler(service *StockService) *StockHandler {
	return &StockHandler{service: service}
}

type CreateStockRequest struct {
	PartCode   string  `json:"stc_part_code" validate:"required"`
	PartOrg    string  `json:"stc_part_org" validate:"required"`
	StoreCode  string  `json:"stc_store_code" validate:"required"`
	StoreOrg   string  `json:"stc_store_org" validate:"required"`
	MinStock   float64 `json:"stc_min_stock"`
	ReorderQty float64 `json:"stc_reorder_qty"`
	ActualQty  float64 `json:"stc_actual_qty"`
}

type UpdateStockRequest struct {
	MinStock   float64 `json:"stc_min_stock"`
	ReorderQty float64 `json:"stc_reorder_qty"`
	ActualQty  float64 `json:"stc_actual_qty"`
}

type AdjustQuantityRequest struct {
	Quantity   float64 `json:"quantity" validate:"required"`
	Adjustment float64 `json:"adjustment"`
}

type CreateBinStockRequest struct {
	PartCode  string  `json:"bis_part_code" validate:"required"`
	PartOrg   string  `json:"bis_part_org" validate:"required"`
	StoreCode string  `json:"bis_store_code" validate:"required"`
	StoreOrg  string  `json:"bis_store_org" validate:"required"`
	BinCode   string  `json:"bis_bin_code" validate:"required"`
	BinOrg    string  `json:"bis_bin_org" validate:"required"`
	Quantity  float64 `json:"bis_quantity"`
}

type UpdateBinStockRequest struct {
	Quantity float64 `json:"bis_quantity"`
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

	var req CreateStockRequest
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

	var req UpdateStockRequest
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

	var req AdjustQuantityRequest
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

	var req CreateBinStockRequest
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

	var req UpdateBinStockRequest
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
