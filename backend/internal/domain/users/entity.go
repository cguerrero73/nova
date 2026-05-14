package users

import "time"

type User struct {
	ID         string    `json:"usr_id"`
	Code       string    `json:"usr_code"`
	Name       string    `json:"usr_name"`
	Email      string    `json:"usr_email"`
	Password   string    `json:"-"`
	Phone      string    `json:"usr_phone"`
	Status     string    `json:"usr_status"`
	DefaultOrg string    `json:"usr_default_org"`
	NotUsed    *string   `json:"usr_notused,omitempty"`
	TenantID   string    `json:"usr_tenant_id"`
	CreatedAt  time.Time `json:"usr_created_at"`
	UpdatedAt  time.Time `json:"usr_updated_at"`
	CreatedBy  string    `json:"usr_created_by,omitempty"`
	UpdatedBy  string    `json:"usr_updated_by,omitempty"`
}

func (u *User) IsActive() bool {
	return u.Status == "ACT" && (u.NotUsed == nil || *u.NotUsed != "+")
}

func (u *User) IsAdmin() bool {
	// Admin check would be done via roles
	return false
}