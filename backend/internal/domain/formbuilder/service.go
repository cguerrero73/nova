package formbuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
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

// ArchiveForm soft-archives a form.
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
