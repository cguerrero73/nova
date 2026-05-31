package users

import "time"

type User struct {
	ID         string    `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   string    `json:"-"`
	Phone      *string   `json:"phone,omitempty"`
	Status     string    `json:"status"`
	DefaultOrg *string   `json:"defaultOrg,omitempty"`
	NotUsed    *string   `json:"notUsed,omitempty"`
	TenantID   *string   `json:"tenantId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	CreatedBy  *string   `json:"createdBy,omitempty"`
	UpdatedBy  *string   `json:"updatedBy,omitempty"`
}

func (u *User) IsActive() bool {
	return u.Status == "ACT" && (u.NotUsed == nil || *u.NotUsed != "+")
}

func (u *User) IsAdmin() bool {
	// Admin check would be done via roles
	return false
}
