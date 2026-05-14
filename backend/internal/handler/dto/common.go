package dto

// Pagination request
type PaginationQuery struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

func (p *PaginationQuery) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	return (p.Page - 1) * p.PageSize
}

func (p *PaginationQuery) GetLimit() int {
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	return p.PageSize
}

// Pagination response
type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewPaginationMeta(page, pageSize, total int) PaginationMeta {
	pages := total / pageSize
	if total%pageSize > 0 {
		pages++
	}
	return PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: pages,
	}
}

// List response wrapper
type ListResponse struct {
	Success bool           `json:"success"`
	Data    interface{}    `json:"data"`
	Meta    PaginationMeta `json:"meta,omitempty"`
}

func NewListResponse(data interface{}, meta PaginationMeta) ListResponse {
	return ListResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

// Error response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}

// Success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func NewSuccessResponse(data interface{}) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Data:    data,
	}
}

func NewMessageResponse(message string) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: message,
	}
}