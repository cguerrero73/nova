package formbuilder_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nova/backend/internal/domain/formbuilder"
)

// --- Mock repositories ---

type mockFormRepo struct {
	forms  map[string]*formbuilder.Form
	nextID int64
}

func newMockFormRepo() *mockFormRepo {
	return &mockFormRepo{forms: make(map[string]*formbuilder.Form), nextID: 1}
}

func (m *mockFormRepo) Create(_ context.Context, form *formbuilder.Form) error {
	if _, exists := m.forms[form.Key]; exists {
		return formbuilder.ErrFormKeyExists
	}
	form.ID = m.nextID
	m.nextID++
	m.forms[form.Key] = form
	return nil
}

func (m *mockFormRepo) FindByKey(_ context.Context, key string) (*formbuilder.Form, error) {
	f, ok := m.forms[key]
	if !ok {
		return nil, nil
	}
	return f, nil
}

func (m *mockFormRepo) List(_ context.Context, status string) ([]*formbuilder.Form, error) {
	var result []*formbuilder.Form
	for _, f := range m.forms {
		if status == "" || f.Status == status {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *mockFormRepo) Archive(_ context.Context, formID int64) error {
	for _, f := range m.forms {
		if f.ID == formID {
			f.Status = "archived"
			return nil
		}
	}
	return nil
}

type mockLayoutRepo struct {
	layouts map[int64]map[string]*formbuilder.Layout // formID → name → layout
	nextID  int64
}

func newMockLayoutRepo() *mockLayoutRepo {
	return &mockLayoutRepo{layouts: make(map[int64]map[string]*formbuilder.Layout), nextID: 1}
}

func (m *mockLayoutRepo) Create(_ context.Context, layout *formbuilder.Layout) error {
	if m.layouts[layout.FormID] == nil {
		m.layouts[layout.FormID] = make(map[string]*formbuilder.Layout)
	}
	if _, exists := m.layouts[layout.FormID][layout.Name]; exists {
		return formbuilder.ErrLayoutNameExists
	}
	layout.ID = m.nextID
	m.nextID++
	m.layouts[layout.FormID][layout.Name] = layout
	return nil
}

func (m *mockLayoutRepo) FindByID(_ context.Context, id int64) (*formbuilder.Layout, error) {
	for _, byName := range m.layouts {
		for _, l := range byName {
			if l.ID == id {
				return l, nil
			}
		}
	}
	return nil, nil
}

func (m *mockLayoutRepo) FindByFormAndName(_ context.Context, formID int64, name string) (*formbuilder.Layout, error) {
	if byName, ok := m.layouts[formID]; ok {
		if l, ok := byName[name]; ok {
			return l, nil
		}
	}
	return nil, nil
}

func (m *mockLayoutRepo) ListByFormID(_ context.Context, formID int64) ([]*formbuilder.Layout, error) {
	var result []*formbuilder.Layout
	if byName, ok := m.layouts[formID]; ok {
		for _, l := range byName {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockLayoutRepo) UpdatePointers(_ context.Context, layoutID int64, draftVersionID, publishedVersionID *int64) error {
	for _, byName := range m.layouts {
		for _, l := range byName {
			if l.ID == layoutID {
				l.DraftVersionID = draftVersionID
				l.PublishedVersionID = publishedVersionID
				return nil
			}
		}
	}
	return nil
}

func (m *mockLayoutRepo) Archive(_ context.Context, layoutID int64) error {
	for _, byName := range m.layouts {
		for _, l := range byName {
			if l.ID == layoutID {
				l.Status = "archived"
				return nil
			}
		}
	}
	return nil
}

type mockVersionRepo struct {
	versions map[int64]*formbuilder.LayoutVersion
	nextID   int64
}

func newMockVersionRepo() *mockVersionRepo {
	return &mockVersionRepo{versions: make(map[int64]*formbuilder.LayoutVersion), nextID: 1}
}

func (m *mockVersionRepo) Create(_ context.Context, version *formbuilder.LayoutVersion) error {
	version.ID = m.nextID
	m.nextID++
	m.versions[version.ID] = version
	return nil
}

func (m *mockVersionRepo) FindByID(_ context.Context, id int64) (*formbuilder.LayoutVersion, error) {
	v, ok := m.versions[id]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (m *mockVersionRepo) FindMaxVersionNumber(_ context.Context, layoutID int64) (int, error) {
	max := 0
	for _, v := range m.versions {
		if v.LayoutID == layoutID && v.VersionNumber > max {
			max = v.VersionNumber
		}
	}
	return max, nil
}

func (m *mockVersionRepo) FindDraft(_ context.Context, layoutID int64) (*formbuilder.LayoutVersion, error) {
	for _, v := range m.versions {
		if v.LayoutID == layoutID && v.Kind == "draft" {
			return v, nil
		}
	}
	return nil, nil
}

func (m *mockVersionRepo) UpdateDraftDefinition(_ context.Context, versionID int64, definition []byte) error {
	if v, ok := m.versions[versionID]; ok {
		v.Definition = definition
	}
	return nil
}

func (m *mockVersionRepo) ListByLayoutID(_ context.Context, layoutID int64) ([]*formbuilder.LayoutVersion, error) {
	var result []*formbuilder.LayoutVersion
	for _, v := range m.versions {
		if v.LayoutID == layoutID {
			result = append(result, v)
		}
	}
	return result, nil
}

type mockAssignmentRepo struct {
	assignments []*formbuilder.RoleAssignment
	nextID      int64
}

func newMockAssignmentRepo() *mockAssignmentRepo {
	return &mockAssignmentRepo{nextID: 1}
}

func (m *mockAssignmentRepo) Create(_ context.Context, assignment *formbuilder.RoleAssignment) error {
	assignment.ID = m.nextID
	m.nextID++
	m.assignments = append(m.assignments, assignment)
	return nil
}

func (m *mockAssignmentRepo) FindActiveByFormAndRole(_ context.Context, formID int64, roleName string) (*formbuilder.RoleAssignment, error) {
	for _, a := range m.assignments {
		if a.FormID == formID && a.RoleName == roleName && a.RevokedAt == nil {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockAssignmentRepo) ListByFormID(_ context.Context, formID int64) ([]*formbuilder.RoleAssignment, error) {
	var result []*formbuilder.RoleAssignment
	for _, a := range m.assignments {
		if a.FormID == formID && a.RevokedAt == nil {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAssignmentRepo) Revoke(_ context.Context, assignmentID int64) error {
	for _, a := range m.assignments {
		if a.ID == assignmentID {
			now := time.Now()
			a.RevokedAt = &now
			return nil
		}
	}
	return nil
}

type mockAuditRepo struct {
	entries []*formbuilder.AuditEntry
}

func newMockAuditRepo() *mockAuditRepo {
	return &mockAuditRepo{}
}

func (m *mockAuditRepo) Create(_ context.Context, entry *formbuilder.AuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAuditRepo) ListByForm(_ context.Context, _ int64, _ string, _ string, _, _ int) ([]*formbuilder.AuditEntry, int, error) {
	return m.entries, len(m.entries), nil
}

// --- Helper ---

func newTestService() (*formbuilder.FormService, *mockFormRepo, *mockLayoutRepo, *mockVersionRepo, *mockAssignmentRepo, *mockAuditRepo) {
	fr := newMockFormRepo()
	lr := newMockLayoutRepo()
	vr := newMockVersionRepo()
	ar := newMockAssignmentRepo()
	au := newMockAuditRepo()
	svc := formbuilder.NewFormService(fr, lr, vr, ar, au)
	return svc, fr, lr, vr, ar, au
}

// --- Tests ---

func TestCreateForm_AutoCreatesDefaultLayout(t *testing.T) {
	svc, _, lr, _, _, au := newTestService()
	ctx := context.Background()

	form, layout, err := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "customer-intake",
		Name: "Customer Intake",
	}, "admin")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form.Key != "customer-intake" {
		t.Errorf("expected form key 'customer-intake', got '%s'", form.Key)
	}
	if layout.Name != "default" {
		t.Errorf("expected default layout, got '%s'", layout.Name)
	}
	if layout.Status != "active" {
		t.Errorf("expected active status, got '%s'", layout.Status)
	}

	// Verify the layout is findable via the repo
	found, _ := lr.FindByFormAndName(ctx, form.ID, "default")
	if found == nil {
		t.Error("default layout not found in repository")
	}

	// Verify audit entries
	if len(au.entries) != 2 {
		t.Fatalf("expected 2 audit entries, got %d", len(au.entries))
	}
	if au.entries[0].Action != "form.create" {
		t.Errorf("expected first audit action 'form.create', got '%s'", au.entries[0].Action)
	}
	if au.entries[1].Action != "layout.create" {
		t.Errorf("expected second audit action 'layout.create', got '%s'", au.entries[1].Action)
	}
	// Check auto_created metadata
	var meta map[string]interface{}
	if err := json.Unmarshal(au.entries[1].Metadata, &meta); err != nil {
		t.Fatalf("failed to unmarshal metadata: %v", err)
	}
	if meta["auto_created"] != true {
		t.Errorf("expected auto_created=true in metadata, got %v", meta["auto_created"])
	}
}

func TestCreateForm_DuplicateKeyRejected(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	_, _, err := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "duplicate",
		Name: "First",
	}, "admin")
	if err != nil {
		t.Fatalf("first create should succeed: %v", err)
	}

	_, _, err = svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "duplicate",
		Name: "Second",
	}, "admin")
	if err == nil {
		t.Fatal("expected error for duplicate key")
	}
}

func TestCreateLayout_RejectsReservedName(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	// Create form first
	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	_, err := svc.CreateLayout(ctx, form.Key, &formbuilder.CreateLayoutRequest{
		Name:        "default",
		DisplayName: "Should Fail",
	}, "admin")

	if err == nil {
		t.Fatal("expected ErrReservedLayoutName")
	}
	if err != formbuilder.ErrReservedLayoutName {
		t.Errorf("expected ErrReservedLayoutName, got %v", err)
	}
}

func TestCreateLayout_Success(t *testing.T) {
	svc, _, lr, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	layout, err := svc.CreateLayout(ctx, form.Key, &formbuilder.CreateLayoutRequest{
		Name:        "admin-full",
		DisplayName: "Admin Full",
	}, "admin")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if layout.Name != "admin-full" {
		t.Errorf("expected layout name 'admin-full', got '%s'", layout.Name)
	}

	// Verify it's in the repo
	found, _ := lr.FindByFormAndName(ctx, form.ID, "admin-full")
	if found == nil {
		t.Error("layout not found in repository")
	}
}

func TestResolve_UnassignedRoleFallsBackToDefault(t *testing.T) {
	svc, _, lr, vr, _, _ := newTestService()
	ctx := context.Background()

	// Create form (auto-creates default layout)
	form, defaultLayout, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "customer-intake",
		Name: "Customer Intake",
	}, "admin")

	// Create a published version for the default layout
	defJSON := json.RawMessage(`{"sections":[{"fields":[{"type":"text","label":"Name"}]}]}`)
	version := &formbuilder.LayoutVersion{
		LayoutID:      defaultLayout.ID,
		VersionNumber: 1,
		Kind:          "published",
		Description:   "Initial publish",
		Definition:    defJSON,
		CreatedBy:     "admin",
		CreatedAt:     time.Now(),
	}
	_ = vr.Create(ctx, version)

	// Point the default layout to the published version
	_ = lr.UpdatePointers(ctx, defaultLayout.ID, nil, &version.ID)

	// Resolve for an unassigned role → should fall back to default
	result, err := svc.Resolve(ctx, form.Key, "viewer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LayoutName != "default" {
		t.Errorf("expected layout 'default', got '%s'", result.LayoutName)
	}
	if result.Version != 1 {
		t.Errorf("expected version 1, got %d", result.Version)
	}
	if string(result.Definition) != string(defJSON) {
		t.Errorf("definition mismatch")
	}
}

func TestResolve_AssignedRoleGetsAssignedLayout(t *testing.T) {
	svc, _, lr, vr, ar, _ := newTestService()
	ctx := context.Background()

	// Create form
	form, defaultLayout, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "customer-intake",
		Name: "Customer Intake",
	}, "admin")

	// Create admin-full layout
	adminLayout, _ := svc.CreateLayout(ctx, form.Key, &formbuilder.CreateLayoutRequest{
		Name:        "admin-full",
		DisplayName: "Admin Full",
	}, "admin")

	// Publish both layouts
	defJSON := json.RawMessage(`{"type":"default"}`)
	adminJSON := json.RawMessage(`{"type":"admin"}`)

	defVer := &formbuilder.LayoutVersion{
		LayoutID: defaultLayout.ID, VersionNumber: 1, Kind: "published",
		Definition: defJSON, CreatedBy: "admin", CreatedAt: time.Now(),
	}
	_ = vr.Create(ctx, defVer)
	_ = lr.UpdatePointers(ctx, defaultLayout.ID, nil, &defVer.ID)

	adminVer := &formbuilder.LayoutVersion{
		LayoutID: adminLayout.ID, VersionNumber: 1, Kind: "published",
		Definition: adminJSON, CreatedBy: "admin", CreatedAt: time.Now(),
	}
	_ = vr.Create(ctx, adminVer)
	_ = lr.UpdatePointers(ctx, adminLayout.ID, nil, &adminVer.ID)

	// Assign admin role → admin-full
	_ = ar.Create(ctx, &formbuilder.RoleAssignment{
		FormID:     form.ID,
		LayoutID:   adminLayout.ID,
		RoleName:   "admin",
		AssignedAt: time.Now(),
		AssignedBy: "admin",
	})

	// Resolve for admin → should get admin-full
	result, err := svc.Resolve(ctx, form.Key, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LayoutName != "admin-full" {
		t.Errorf("expected layout 'admin-full', got '%s'", result.LayoutName)
	}
	if string(result.Definition) != string(adminJSON) {
		t.Errorf("expected admin definition, got %s", string(result.Definition))
	}

	// Resolve for viewer → should fall back to default
	result, err = svc.Resolve(ctx, form.Key, "viewer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LayoutName != "default" {
		t.Errorf("expected layout 'default', got '%s'", result.LayoutName)
	}
}

func TestResolve_NoPublishedVersion_ReturnsNotPublished(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	// Create form (default layout has no published version)
	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "empty-form",
		Name: "Empty",
	}, "admin")

	_, err := svc.Resolve(ctx, form.Key, "viewer")
	if err == nil {
		t.Fatal("expected error for no published version")
	}
	if err != formbuilder.ErrFormLayoutNotPublished {
		t.Errorf("expected ErrFormLayoutNotPublished, got %v", err)
	}
}

func TestResolve_ArchivedForm_ReturnsArchived(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "archived-form",
		Name: "Archived",
	}, "admin")
	_ = svc.ArchiveForm(ctx, form.Key, "admin")

	_, err := svc.Resolve(ctx, form.Key, "viewer")
	if err == nil {
		t.Fatal("expected error for archived form")
	}
	if err != formbuilder.ErrFormArchived {
		t.Errorf("expected ErrFormArchived, got %v", err)
	}
}

func TestListForms_Empty(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	forms, err := svc.ListForms(ctx, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(forms) != 0 {
		t.Errorf("expected 0 forms, got %d", len(forms))
	}
}

func TestListLayouts_IncludesDefault(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test",
		Name: "Test",
	}, "admin")

	layouts, err := svc.ListLayouts(ctx, form.Key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(layouts) != 1 {
		t.Fatalf("expected 1 layout (default), got %d", len(layouts))
	}
	if layouts[0].Name != "default" {
		t.Errorf("expected 'default', got '%s'", layouts[0].Name)
	}
}
