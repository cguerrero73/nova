package formbuilder

import "context"

// FormRepository manages form definition persistence.
type FormRepository interface {
	// Create inserts a new form definition. Returns the form with ID populated.
	Create(ctx context.Context, form *Form) error

	// FindByKey loads a form by its unique key.
	FindByKey(ctx context.Context, key string) (*Form, error)

	// List returns all forms (optionally filtered by status).
	List(ctx context.Context, status string) ([]*Form, error)

	// Archive sets frm_status to 'archived'.
	Archive(ctx context.Context, formID int64) error
}

// LayoutRepository manages layout persistence within a form.
type LayoutRepository interface {
	// Create inserts a new layout. Returns the layout with ID populated.
	Create(ctx context.Context, layout *Layout) error

	// FindByID loads a layout by its ID.
	FindByID(ctx context.Context, id int64) (*Layout, error)

	// FindByFormAndName loads a layout by form ID and name.
	FindByFormAndName(ctx context.Context, formID int64, name string) (*Layout, error)

	// ListByFormID returns all layouts for a form.
	ListByFormID(ctx context.Context, formID int64) ([]*Layout, error)

	// UpdatePointers updates draft/published version pointers.
	UpdatePointers(ctx context.Context, layoutID int64, draftVersionID, publishedVersionID *int64) error

	// Archive sets fl_status to 'archived'.
	Archive(ctx context.Context, layoutID int64) error

	// ArchiveByFormID archives all layouts for a form (cascade from form archive).
	ArchiveByFormID(ctx context.Context, formID int64) error
}

// LayoutVersionRepository manages version snapshot persistence.
type LayoutVersionRepository interface {
	// Create inserts a new version row. Returns the version with ID populated.
	Create(ctx context.Context, version *LayoutVersion) error

	// FindByID loads a version by ID.
	FindByID(ctx context.Context, id int64) (*LayoutVersion, error)

	// FindMaxVersionNumber returns the highest version_number for a layout.
	FindMaxVersionNumber(ctx context.Context, layoutID int64) (int, error)

	// FindDraft returns the current draft version for a layout (if any).
	FindDraft(ctx context.Context, layoutID int64) (*LayoutVersion, error)

	// UpdateDraftDefinition updates the JSON definition of an existing draft.
	UpdateDraftDefinition(ctx context.Context, versionID int64, definition []byte) error

	// UpdateKind changes the kind of a version row (e.g., draft → archived during publish).
	UpdateKind(ctx context.Context, versionID int64, newKind string) error

	// FindByLayoutAndVersionNumber loads a specific version by layout ID and version number.
	FindByLayoutAndVersionNumber(ctx context.Context, layoutID int64, versionNumber int) (*LayoutVersion, error)

	// ListByLayoutID returns all versions for a layout ordered by version_number DESC.
	ListByLayoutID(ctx context.Context, layoutID int64) ([]*LayoutVersion, error)
}

// AssignmentRepository manages role-to-layout assignment persistence.
type AssignmentRepository interface {
	// Create inserts a new assignment.
	Create(ctx context.Context, assignment *RoleAssignment) error

	// FindActiveByFormAndRole loads the active (non-revoked) assignment for a form+role.
	FindActiveByFormAndRole(ctx context.Context, formID int64, roleName string) (*RoleAssignment, error)

	// ListByFormID returns all assignments for a form.
	ListByFormID(ctx context.Context, formID int64) ([]*RoleAssignment, error)

	// Revoke sets fra_revoked_at on the active assignment.
	Revoke(ctx context.Context, assignmentID int64) error
}

// AuditLogRepository manages audit log persistence.
type AuditLogRepository interface {
	// Create inserts a new audit entry.
	Create(ctx context.Context, entry *AuditEntry) error

	// ListByForm returns audit entries for a form, with optional filters and pagination.
	ListByForm(ctx context.Context, formID int64, action string, entityType string, limit, offset int) ([]*AuditEntry, int, error)
}
