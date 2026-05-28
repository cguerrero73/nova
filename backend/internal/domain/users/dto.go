package users

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Code       string `json:"usr_code"`
	Name       string `json:"usr_name" validate:"required"`
	Email      string `json:"usr_email" validate:"required,email"`
	Password   string `json:"usr_password" validate:"required,min=8"`
	Phone      string `json:"usr_phone"`
	DefaultOrg string `json:"usr_default_org"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Name       string `json:"usr_name"`
	Email      string `json:"usr_email" validate:"email"`
	Phone      string `json:"usr_phone"`
	Status     string `json:"usr_status"`
	DefaultOrg string `json:"usr_default_org"`
}
