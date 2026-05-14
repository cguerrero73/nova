package events

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/internal/infrastructure/middleware"
	"github.com/nova/backend/pkg/errors"
)

type EventHandler struct {
	service *EventService
}

func NewEventHandler(service *EventService) *EventHandler {
	return &EventHandler{service: service}
}

type CreateEventRequest struct {
	Code      string `json:"evt_code" validate:"required"`
	Org       string `json:"evt_org" validate:"required"`
	Desc      string `json:"evt_desc"`
	Type      string `json:"evt_type"`
	RType     string `json:"evt_rtype"`
	Status    string `json:"evt_status"`
	RStatus   string `json:"evt_rstatus"`
	Object    string `json:"evt_object"`
	ObjectOrg string `json:"evt_object_org"`
}

type UpdateEventRequest struct {
	Org       string `json:"evt_org"`
	Desc      string `json:"evt_desc"`
	Type      string `json:"evt_type"`
	RType     string `json:"evt_rtype"`
	Status    string `json:"evt_status"`
	RStatus   string `json:"evt_rstatus"`
	Object    string `json:"evt_object"`
	ObjectOrg string `json:"evt_object_org"`
}

type UpdateStatusRequest struct {
	Status string `json:"evt_status" validate:"required"`
}

func (h *EventHandler) List(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	org := c.Query("org")
	pagination := dto.PaginationQuery{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 20),
	}

	events, total, err := h.service.FindAll(c.Context(), tenant, org, pagination.GetLimit(), pagination.GetOffset())
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewListResponse(events, dto.NewPaginationMeta(pagination.Page, pagination.PageSize, total)))
}

func (h *EventHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	event, err := h.service.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	if event == nil {
		return c.Status(404).JSON(dto.NewErrorResponse("NOT_FOUND", "Event not found"))
	}

	return c.JSON(dto.NewSuccessResponse(event))
}

func (h *EventHandler) Create(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	var req CreateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	event, err := h.service.Create(c.Context(), tenant, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(event))
}

func (h *EventHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	event, err := h.service.Update(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(event))
}

func (h *EventHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("Event deleted successfully"))
}

func (h *EventHandler) GetByObject(c *fiber.Ctx) error {
	code := c.Params("code")
	org := c.Params("org")

	events, err := h.service.FindByObject(c.Context(), code, org)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(events))
}

func (h *EventHandler) UpdateStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	event, err := h.service.UpdateStatus(c.Context(), id, req.Status)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(event))
}
