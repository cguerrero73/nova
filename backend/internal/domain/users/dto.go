package users

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Code       string `json:"code"`
	Name       string `json:"name" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8"`
	Phone      string `json:"phone"`
	DefaultOrg string `json:"defaultOrg"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email" validate:"email"`
	Phone      string `json:"phone"`
	Status     string `json:"status"`
	DefaultOrg string `json:"defaultOrg"`
}
