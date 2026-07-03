package formbuilder

import "encoding/json"

// CreateFormRequest is the HTTP request body for creating a form.
type CreateFormRequest struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateLayoutRequest is the HTTP request body for creating a layout.
type CreateLayoutRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

// PublishRequest is the HTTP request body for publishing a draft.
type PublishRequest struct {
	Description string `json:"description"`
}

// AssignRoleRequest is the HTTP request body for assigning a layout to a role.
type AssignRoleRequest struct {
	LayoutName string `json:"layoutName"`
}

// RevertRequest is the HTTP request body for reverting to a previous version.
type RevertRequest struct {
	VersionNumber int `json:"versionNumber"`
}

// FormResponse is the HTTP response for a single form.
type FormResponse struct {
	ID          int64  `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedBy   string `json:"createdBy"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// LayoutResponse is the HTTP response for a single layout.
type LayoutResponse struct {
	ID                 int64  `json:"id"`
	FormID             int64  `json:"formId"`
	Name               string `json:"name"`
	DisplayName        string `json:"displayName"`
	Description        string `json:"description"`
	Status             string `json:"status"`
	DraftVersionID     *int64 `json:"draftVersionId"`
	PublishedVersionID *int64 `json:"publishedVersionId"`
	CreatedBy          string `json:"createdBy"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

// LayoutVersionResponse is the HTTP response for a version.
type LayoutVersionResponse struct {
	ID            int64           `json:"id"`
	LayoutID      int64           `json:"layoutId"`
	VersionNumber int             `json:"versionNumber"`
	Kind          string          `json:"kind"`
	Description   string          `json:"description"`
	Definition    json.RawMessage `json:"definition"`
	CreatedBy     string          `json:"createdBy"`
	CreatedAt     string          `json:"createdAt"`
	PublishedAt   *string         `json:"publishedAt"`
}

// AssignmentResponse is the HTTP response for a role assignment.
type AssignmentResponse struct {
	ID         int64  `json:"id"`
	FormID     int64  `json:"formId"`
	LayoutID   int64  `json:"layoutId"`
	LayoutName string `json:"layoutName"`
	RoleName   string `json:"roleName"`
	AssignedAt string `json:"assignedAt"`
	AssignedBy string `json:"assignedBy"`
}

// ResolveResponse is the HTTP response for runtime resolution.
type ResolveResponse struct {
	FormKey    string          `json:"formKey"`
	LayoutName string          `json:"layoutName"`
	Version    int             `json:"version"`
	Definition json.RawMessage `json:"definition"`
}
