package errors

import "fmt"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

var (
	ErrNotFound     = New("NOT_FOUND", "Resource not found", 404)
	ErrUnauthorized = New("UNAUTHORIZED", "Unauthorized", 401)
	ErrForbidden    = New("FORBIDDEN", "Access denied", 403)
	ErrBadRequest   = New("BAD_REQUEST", "Invalid request", 400)
	ErrInternal     = New("INTERNAL", "Internal server error", 500)
)

func ErrInvalidCredentials() *AppError {
	return New("INVALID_CREDENTIALS", "Invalid email or password", 401)
}

func ErrTenantRequired() *AppError {
	return New("TENANT_REQUIRED", "Tenant code is required", 400)
}

func ErrUserNotFound() *AppError {
	return New("USER_NOT_FOUND", "User not found", 404)
}

func ErrUserExists() *AppError {
	return New("USER_EXISTS", "User with this email already exists", 409)
}

func ErrInvalidToken() *AppError {
	return New("INVALID_TOKEN", "Invalid or expired token", 401)
}

func ErrOrgRequired() *AppError {
	return New("ORG_REQUIRED", "Organization is required", 400)
}

func ErrNotSuperAdmin() *AppError {
	return New("NOT_SUPERADMIN", "This action requires superadmin privileges", 403)
}
