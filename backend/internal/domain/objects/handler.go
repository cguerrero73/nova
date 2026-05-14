package objects

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nova/backend/internal/handler/dto"
	"github.com/nova/backend/internal/infrastructure/middleware"
	"github.com/nova/backend/pkg/errors"
)

type ObjectHandler struct {
	service *ObjectService
}

func NewObjectHandler(service *ObjectService) *ObjectHandler {
	return &ObjectHandler{service: service}
}

type CreateObjectRequest struct {
	Code        string `json:"obj_code" validate:"required"`
	Type        string `json:"obj_type"`
	Desc        string `json:"obj_desc"`
	Serial      string `json:"obj_serial"`
	Status      string `json:"obj_status"`
	Org         string `json:"obj_org" validate:"required"`
	ParentCode  string `json:"obj_parent_code"`
	ParentOrg   string `json:"obj_parent_org"`
	InstallDate string `json:"obj_install_date"`
}

type UpdateObjectRequest struct {
	Type        string `json:"obj_type"`
	Desc        string `json:"obj_desc"`
	Serial      string `json:"obj_serial"`
	Status      string `json:"obj_status"`
	Org         string `json:"obj_org"`
	ParentCode  string `json:"obj_parent_code"`
	ParentOrg   string `json:"obj_parent_org"`
	InstallDate string `json:"obj_install_date"`
}

func (h *ObjectHandler) List(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	org := c.Query("org")
	pagination := dto.PaginationQuery{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 20),
	}

	objects, total, err := h.service.FindAll(c.Context(), tenant, org, pagination.GetLimit(), pagination.GetOffset())
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewListResponse(objects, dto.NewPaginationMeta(pagination.Page, pagination.PageSize, total)))
}

func (h *ObjectHandler) Get(c *fiber.Ctx) error {
	id := c.Params("id")
	obj, err := h.service.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}
	if obj == nil {
		return c.Status(404).JSON(dto.NewErrorResponse("NOT_FOUND", "Object not found"))
	}

	return c.JSON(dto.NewSuccessResponse(obj))
}

func (h *ObjectHandler) Create(c *fiber.Ctx) error {
	tenant := middleware.GetTenant(c)
	if tenant == "" {
		return c.Status(400).JSON(errors.ErrTenantRequired())
	}

	var req CreateObjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	obj, err := h.service.Create(c.Context(), tenant, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.Status(201).JSON(dto.NewSuccessResponse(obj))
}

func (h *ObjectHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateObjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(errors.ErrBadRequest)
	}

	obj, err := h.service.Update(c.Context(), id, &req)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return c.Status(appErr.Status).JSON(appErr)
		}
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(obj))
}

func (h *ObjectHandler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.service.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewMessageResponse("Object deleted successfully"))
}

func (h *ObjectHandler) GetChildren(c *fiber.Ctx) error {
	parentCode := c.Query("parent_code")
	parentOrg := c.Query("parent_org")

	if parentCode == "" || parentOrg == "" {
		return c.Status(400).JSON(dto.NewErrorResponse("BAD_REQUEST", "parent_code and parent_org are required"))
	}

	children, err := h.service.FindChildren(c.Context(), parentCode, parentOrg)
	if err != nil {
		return c.Status(500).JSON(dto.NewErrorResponse("INTERNAL", err.Error()))
	}

	return c.JSON(dto.NewSuccessResponse(children))
}
