package formbuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	apperrors "github.com/nova/backend/pkg/errors"
)

// FormService encapsulates form builder business logic with default-layout invariants.
type FormService struct {
	forms       FormRepository
	layouts     LayoutRepository
	versions    LayoutVersionRepository
	assignments AssignmentRepository
	audit       AuditLogRepository
}

// NewFormService creates a new FormService with all required repositories.
func NewFormService(
	forms FormRepository,
	layouts LayoutRepository,
	versions LayoutVersionRepository,
	assignments AssignmentRepository,
	audit AuditLogRepository,
) *FormService {
	return &FormService{
		forms:       forms,
		layouts:     layouts,
		versions:    versions,
		assignments: assignments,
		audit:       audit,
	}
}

// CreateForm creates a form and auto-creates the "default" layout in the same
// logical operation. The caller (handler) is responsible for wrapping this in
// a transaction if atomicity is required at the DB level.
func (s *FormService) CreateForm(ctx context.Context, req *CreateFormRequest, actor string) (*Form, *Layout, error) {
	now := time.Now()

	form := &Form{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Status:      "active",
		CreatedBy:   actor,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.forms.Create(ctx, form); err != nil {
		return nil, nil, fmt.Errorf("creating form: %w", err)
	}

	// Auto-create the "default" layout
	defaultLayout := &Layout{
		FormID:      form.ID,
		Name:        "default",
		DisplayName: "Default",
		Description: "Auto-created default layout",
		Status:      "active",
		CreatedBy:   actor,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.layouts.Create(ctx, defaultLayout); err != nil {
		return nil, nil, fmt.Errorf("creating default layout: %w", err)
	}

	// Write audit entries
	s.writeAudit(ctx, actor, "form.create", "form", form.ID, json.RawMessage(`{}`))
	s.writeAudit(ctx, actor, "layout.create", "layout", defaultLayout.ID, json.RawMessage(`{"auto_created":true}`))

	return form, defaultLayout, nil
}

// GetForm loads a form by its key.
func (s *FormService) GetForm(ctx context.Context, formKey string) (*Form, error) {
	form, err := s.forms.FindByKey(ctx, formKey)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, ErrFormNotFound
	}
	return form, nil
}

// ListForms returns all forms, optionally filtered by status.
func (s *FormService) ListForms(ctx context.Context, status string) ([]*Form, error) {
	return s.forms.List(ctx, status)
}

// ArchiveForm soft-archives a form and cascades to all its layouts.
func (s *FormService) ArchiveForm(ctx context.Context, formKey string, actor string) error {
	form, err := s.forms.FindByKey(ctx, formKey)
	if err != nil {
		return err
	}
	if form == nil {
		return ErrFormNotFound
	}
	if form.Status == "archived" {
		return ErrFormArchived
	}

	if err := s.forms.Archive(ctx, form.ID); err != nil {
		return fmt.Errorf("archiving form: %w", err)
	}

	// Cascade: archive all layouts
	if err := s.layouts.ArchiveByFormID(ctx, form.ID); err != nil {
		return fmt.Errorf("cascading archive to layouts: %w", err)
	}

	s.writeAudit(ctx, actor, "form.archive", "form", form.ID, json.RawMessage(`{}`))
	return nil
}

// CreateLayout creates a new named layout within a form.
// Rejects the reserved name "default" with ErrReservedLayoutName.
func (s *FormService) CreateLayout(ctx context.Context, formKey string, req *CreateLayoutRequest, actor string) (*Layout, error) {
	if req.Name == "default" {
		return nil, ErrReservedLayoutName
	}

	form, err := s.forms.FindByKey(ctx, formKey)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, ErrFormNotFound
	}
	if form.Status == "archived" {
		return nil, ErrFormArchived
	}

	// Check for duplicate name
	existing, err := s.layouts.FindByFormAndName(ctx, form.ID, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrLayoutNameExists
	}

	now := time.Now()
	layout := &Layout{
		FormID:      form.ID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Status:      "active",
		CreatedBy:   actor,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.layouts.Create(ctx, layout); err != nil {
		return nil, fmt.Errorf("creating layout: %w", err)
	}

	s.writeAudit(ctx, actor, "layout.create", "layout", layout.ID, json.RawMessage(`{}`))
	return layout, nil
}

// ListLayouts returns all layouts for a form.
func (s *FormService) ListLayouts(ctx context.Context, formKey string) ([]*Layout, error) {
	form, err := s.forms.FindByKey(ctx, formKey)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, ErrFormNotFound
	}

	return s.layouts.ListByFormID(ctx, form.ID)
}

// GetDraft returns the current draft version of a layout.
func (s *FormService) GetDraft(ctx context.Context, formKey, layoutName string) (*LayoutVersion, error) {
	form, layout, err := s.resolveFormLayout(ctx, formKey, layoutName)
	if err != nil {
		return nil, err
	}
	_ = form

	if layout.DraftVersionID == nil {
		return nil, ErrLayoutNotFound
	}

	return s.versions.FindByID(ctx, *layout.DraftVersionID)
}

// ListAssignments returns all assignments for a form.
func (s *FormService) ListAssignments(ctx context.Context, formKey string) ([]*RoleAssignment, error) {
	form, err := s.forms.FindByKey(ctx, formKey)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, ErrFormNotFound
	}

	return s.assignments.ListByFormID(ctx, form.ID)
}

// SaveDraft creates or updates the draft version for a layout.
// If the layout has no draft, a new draft version row is inserted and
// fl_draft_version_id is repointed. If a draft already exists, its
// definition is updated in place.
func (s *FormService) SaveDraft(ctx context.Context, formKey, layoutName string, definition json.RawMessage, actor string) (*LayoutVersion, error) {
	form, layout, err := s.resolveFormLayout(ctx, formKey, layoutName)
	if err != nil {
		return nil, err
	}
	_ = form

	if layout.Status == "archived" {
		return nil, ErrLayoutArchived
	}

	// Validate JSON is well-formed
	if !json.Valid(definition) {
		return nil, apperrors.New("INVALID_JSON", "The layout definition is not valid JSON", 400)
	}

	now := time.Now()

	if layout.DraftVersionID == nil {
		// No draft exists — create one
		maxVer, err := s.versions.FindMaxVersionNumber(ctx, layout.ID)
		if err != nil {
			return nil, fmt.Errorf("finding max version: %w", err)
		}

		draft := &LayoutVersion{
			LayoutID:      layout.ID,
			VersionNumber: maxVer + 1,
			Kind:          "draft",
			Description:   "",
			Definition:    definition,
			CreatedBy:     actor,
			CreatedAt:     now,
		}

		if err := s.versions.Create(ctx, draft); err != nil {
			return nil, fmt.Errorf("creating draft version: %w", err)
		}

		// Repoint the layout's draft pointer
		if err := s.layouts.UpdatePointers(ctx, layout.ID, &draft.ID, layout.PublishedVersionID); err != nil {
			return nil, fmt.Errorf("updating draft pointer: %w", err)
		}

		s.writeAudit(ctx, actor, "version.draft_save", "version", draft.ID, json.RawMessage(`{"action":"create"}`))
		return draft, nil
	}

	// Draft exists — update its definition
	if err := s.versions.UpdateDraftDefinition(ctx, *layout.DraftVersionID, definition); err != nil {
		return nil, fmt.Errorf("updating draft definition: %w", err)
	}

	// Reload the draft to return the updated version
	draft, err := s.versions.FindByID(ctx, *layout.DraftVersionID)
	if err != nil {
		return nil, fmt.Errorf("reloading draft: %w", err)
	}

	s.writeAudit(ctx, actor, "version.draft_save", "version", draft.ID, json.RawMessage(`{"action":"update"}`))
	return draft, nil
}

// Publish promotes the current draft to a published version.
// Requires a non-empty description (commit message). The previous published
// version (if any) remains in the table; the draft pointer is cleared.
func (s *FormService) Publish(ctx context.Context, formKey, layoutName, description, actor string) (*LayoutVersion, error) {
	if description == "" {
		return nil, ErrPublishDescriptionRequired
	}

	form, layout, err := s.resolveFormLayout(ctx, formKey, layoutName)
	if err != nil {
		return nil, err
	}
	_ = form

	if layout.Status == "archived" {
		return nil, ErrLayoutArchived
	}

	if layout.DraftVersionID == nil {
		return nil, ErrNoDraft
	}

	// Load the draft
	draft, err := s.versions.FindByID(ctx, *layout.DraftVersionID)
	if err != nil {
		return nil, fmt.Errorf("loading draft: %w", err)
	}
	if draft == nil {
		return nil, ErrNoDraft
	}

	// Determine the next version number
	maxVer, err := s.versions.FindMaxVersionNumber(ctx, layout.ID)
	if err != nil {
		return nil, fmt.Errorf("finding max version: %w", err)
	}

	now := time.Now()

	// INSERT new published version (copy definition from draft)
	published := &LayoutVersion{
		LayoutID:      layout.ID,
		VersionNumber: maxVer + 1,
		Kind:          "published",
		Description:   description,
		Definition:    draft.Definition,
		CreatedBy:     actor,
		CreatedAt:     now,
		PublishedAt:   &now,
	}

	if err := s.versions.Create(ctx, published); err != nil {
		return nil, fmt.Errorf("creating published version: %w", err)
	}

	// Transition the draft to 'archived' so the one-draft-per-layout unique
	// index is freed and a new draft can be created later.
	if err := s.versions.UpdateKind(ctx, draft.ID, "archived"); err != nil {
		return nil, fmt.Errorf("archiving consumed draft: %w", err)
	}

	// Update layout pointers: clear draft, point to new published
	if err := s.layouts.UpdatePointers(ctx, layout.ID, nil, &published.ID); err != nil {
		return nil, fmt.Errorf("updating published pointer: %w", err)
	}

	meta, _ := json.Marshal(map[string]interface{}{
		"version":     published.VersionNumber,
		"description": description,
	})
	s.writeAudit(ctx, actor, "version.publish", "version", published.ID, meta)

	return published, nil
}

// Revert copies a previous published version's definition back as a new draft.
// If the layout already has a draft, it is updated with the reverted definition.
// If not, a new draft is created.
func (s *FormService) Revert(ctx context.Context, formKey, layoutName string, versionNumber int, actor string) (*LayoutVersion, error) {
	form, layout, err := s.resolveFormLayout(ctx, formKey, layoutName)
	if err != nil {
		return nil, err
	}
	_ = form

	if layout.Status == "archived" {
		return nil, ErrLayoutArchived
	}

	// Load the target version
	sourceVersion, err := s.versions.FindByLayoutAndVersionNumber(ctx, layout.ID, versionNumber)
	if err != nil {
		return nil, fmt.Errorf("finding version %d: %w", versionNumber, err)
	}
	if sourceVersion == nil {
		return nil, ErrVersionNotFound
	}

	now := time.Now()

	if layout.DraftVersionID != nil {
		// Update existing draft with the reverted definition
		if err := s.versions.UpdateDraftDefinition(ctx, *layout.DraftVersionID, sourceVersion.Definition); err != nil {
			return nil, fmt.Errorf("updating draft with reverted definition: %w", err)
		}

		draft, err := s.versions.FindByID(ctx, *layout.DraftVersionID)
		if err != nil {
			return nil, fmt.Errorf("reloading draft: %w", err)
		}

		meta, _ := json.Marshal(map[string]interface{}{
			"from_version": versionNumber,
		})
		s.writeAudit(ctx, actor, "version.revert", "version", draft.ID, meta)
		return draft, nil
	}

	// No draft exists — create one with the reverted definition
	maxVer, err := s.versions.FindMaxVersionNumber(ctx, layout.ID)
	if err != nil {
		return nil, fmt.Errorf("finding max version: %w", err)
	}

	draft := &LayoutVersion{
		LayoutID:      layout.ID,
		VersionNumber: maxVer + 1,
		Kind:          "draft",
		Description:   "",
		Definition:    sourceVersion.Definition,
		CreatedBy:     actor,
		CreatedAt:     now,
	}

	if err := s.versions.Create(ctx, draft); err != nil {
		return nil, fmt.Errorf("creating draft from revert: %w", err)
	}

	if err := s.layouts.UpdatePointers(ctx, layout.ID, &draft.ID, layout.PublishedVersionID); err != nil {
		return nil, fmt.Errorf("updating draft pointer: %w", err)
	}

	meta, _ := json.Marshal(map[string]interface{}{
		"from_version": versionNumber,
	})
	s.writeAudit(ctx, actor, "version.revert", "version", draft.ID, meta)
	return draft, nil
}

// ArchiveLayout soft-archives a layout. The default layout of an active form
// cannot be archived (returns ErrCannotArchiveDefault).
func (s *FormService) ArchiveLayout(ctx context.Context, formKey, layoutName, actor string) error {
	form, layout, err := s.resolveFormLayout(ctx, formKey, layoutName)
	if err != nil {
		return err
	}

	if layout.Status == "archived" {
		return nil // idempotent
	}

	// Invariant: default layout cannot be archived while form is active
	if layout.Name == "default" && form.Status == "active" {
		return ErrCannotArchiveDefault
	}

	if err := s.layouts.Archive(ctx, layout.ID); err != nil {
		return fmt.Errorf("archiving layout: %w", err)
	}

	s.writeAudit(ctx, actor, "layout.archive", "layout", layout.ID, json.RawMessage(`{}`))
	return nil
}

// AssignRole assigns a layout to a role for a given form. If an active
// assignment already exists for this (form, role), it is revoked first.
func (s *FormService) AssignRole(ctx context.Context, formKey, roleName, layoutName, actor string) (*RoleAssignment, error) {
	form, err := s.forms.FindByKey(ctx, formKey)
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, ErrFormNotFound
	}
	if form.Status == "archived" {
		return nil, ErrFormArchived
	}

	// Verify the layout belongs to this form
	layout, err := s.layouts.FindByFormAndName(ctx, form.ID, layoutName)
	if err != nil {
		return nil, err
	}
	if layout == nil {
		return nil, ErrLayoutNotFound
	}
	if layout.Status == "archived" {
		return nil, ErrLayoutArchived
	}

	// Revoke existing assignment if any (replace semantics)
	existing, err := s.assignments.FindActiveByFormAndRole(ctx, form.ID, roleName)
	if err != nil {
		return nil, fmt.Errorf("checking existing assignment: %w", err)
	}
	if existing != nil {
		if err := s.assignments.Revoke(ctx, existing.ID); err != nil {
			return nil, fmt.Errorf("revoking previous assignment: %w", err)
		}
	}

	now := time.Now()
	assignment := &RoleAssignment{
		FormID:     form.ID,
		LayoutID:   layout.ID,
		RoleName:   roleName,
		AssignedAt: now,
		AssignedBy: actor,
	}

	if err := s.assignments.Create(ctx, assignment); err != nil {
		return nil, fmt.Errorf("creating assignment: %w", err)
	}

	meta, _ := json.Marshal(map[string]interface{}{
		"layout_name": layoutName,
		"role_name":   roleName,
	})
	s.writeAudit(ctx, actor, "layout.assign", "assignment", assignment.ID, meta)

	return assignment, nil
}

// RevokeAssignment soft-revokes the active assignment for a role on a form.
func (s *FormService) RevokeAssignment(ctx context.Context, formKey, roleName, actor string) error {
	form, err := s.forms.FindByKey(ctx, formKey)
	if err != nil {
		return err
	}
	if form == nil {
		return ErrFormNotFound
	}

	existing, err := s.assignments.FindActiveByFormAndRole(ctx, form.ID, roleName)
	if err != nil {
		return fmt.Errorf("finding assignment: %w", err)
	}
	if existing == nil {
		return ErrAssignmentNotFound
	}

	if err := s.assignments.Revoke(ctx, existing.ID); err != nil {
		return fmt.Errorf("revoking assignment: %w", err)
	}

	meta, _ := json.Marshal(map[string]interface{}{
		"role_name": roleName,
	})
	s.writeAudit(ctx, actor, "layout.unassign", "assignment", existing.ID, meta)
	return nil
}

// ListVersions returns all versions for a layout ordered by version_number DESC.
func (s *FormService) ListVersions(ctx context.Context, formKey, layoutName string) ([]*LayoutVersion, error) {
	_, layout, err := s.resolveFormLayout(ctx, formKey, layoutName)
	if err != nil {
		return nil, err
	}

	return s.versions.ListByLayoutID(ctx, layout.ID)
}

// GetVersion returns a specific version by version number for a layout.
func (s *FormService) GetVersion(ctx context.Context, formKey, layoutName string, versionNumber int) (*LayoutVersion, error) {
	_, layout, err := s.resolveFormLayout(ctx, formKey, layoutName)
	if err != nil {
		return nil, err
	}

	version, err := s.versions.FindByLayoutAndVersionNumber(ctx, layout.ID, versionNumber)
	if err != nil {
		return nil, fmt.Errorf("finding version: %w", err)
	}
	if version == nil {
		return nil, ErrVersionNotFound
	}

	return version, nil
}

// resolveFormLayout loads a form and a layout by their keys.
func (s *FormService) resolveFormLayout(ctx context.Context, formKey, layoutName string) (*Form, *Layout, error) {
	form, err := s.forms.FindByKey(ctx, formKey)
	if err != nil {
		return nil, nil, err
	}
	if form == nil {
		return nil, nil, ErrFormNotFound
	}

	layout, err := s.layouts.FindByFormAndName(ctx, form.ID, layoutName)
	if err != nil {
		return nil, nil, err
	}
	if layout == nil {
		return nil, nil, ErrLayoutNotFound
	}

	return form, layout, nil
}

// writeAudit is a helper that writes an audit entry, ignoring errors (best-effort).
func (s *FormService) writeAudit(ctx context.Context, actor, action, entityType string, entityID int64, metadata json.RawMessage) {
	if s.audit == nil {
		return
	}
	entry := &AuditEntry{
		ActorUserID: actor,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
	}
	// Best-effort: log but don't fail the operation
	_ = s.audit.Create(ctx, entry)
}
