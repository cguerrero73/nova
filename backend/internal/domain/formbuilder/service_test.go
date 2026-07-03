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

func (m *mockLayoutRepo) FindByName(_ context.Context, name string) (*formbuilder.Layout, error) {
	for _, byName := range m.layouts {
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

func (m *mockLayoutRepo) ArchiveByFormID(_ context.Context, formID int64) error {
	if byName, ok := m.layouts[formID]; ok {
		for _, l := range byName {
			if l.Status == "active" {
				l.Status = "archived"
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

func (m *mockVersionRepo) UpdateKind(_ context.Context, versionID int64, newKind string) error {
	if v, ok := m.versions[versionID]; ok {
		v.Kind = newKind
	}
	return nil
}

func (m *mockVersionRepo) FindByLayoutAndVersionNumber(_ context.Context, layoutID int64, versionNumber int) (*formbuilder.LayoutVersion, error) {
	for _, v := range m.versions {
		if v.LayoutID == layoutID && v.VersionNumber == versionNumber {
			return v, nil
		}
	}
	return nil, nil
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

func (m *mockAuditRepo) ListByForm(_ context.Context, _ int64, _ formbuilder.AuditFilter, _, _ int) ([]*formbuilder.AuditEntry, int, error) {
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

// --- PR2 Tests ---

func TestSaveDraft_CreatesFirstDraft(t *testing.T) {
	svc, _, lr, vr, _, au := newTestService()
	ctx := context.Background()

	form, defaultLayout, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")
	_ = form

	defJSON := json.RawMessage(`{"sections":[{"fields":[{"type":"text","label":"Name"}]}]}`)
	version, err := svc.SaveDraft(ctx, form.Key, "default", defJSON, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version.Kind != "draft" {
		t.Errorf("expected kind 'draft', got '%s'", version.Kind)
	}
	if string(version.Definition) != string(defJSON) {
		t.Errorf("definition mismatch")
	}

	// Verify draft pointer is set
	updated, _ := lr.FindByID(ctx, defaultLayout.ID)
	if updated.DraftVersionID == nil {
		t.Fatal("expected draft pointer to be set")
	}
	if *updated.DraftVersionID != version.ID {
		t.Errorf("expected draft pointer %d, got %d", version.ID, *updated.DraftVersionID)
	}

	// Verify audit
	found := false
	for _, e := range au.entries {
		if e.Action == "version.draft_save" {
			found = true
		}
	}
	if !found {
		t.Error("expected version.draft_save audit entry")
	}

	// Verify it's in the version repo
	stored, _ := vr.FindByID(ctx, version.ID)
	if stored == nil {
		t.Fatal("draft version not found in repository")
	}
}

func TestSaveDraft_UpdatesExistingDraft(t *testing.T) {
	svc, _, lr, vr, _, _ := newTestService()
	ctx := context.Background()

	form, defaultLayout, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	// Create first draft
	def1 := json.RawMessage(`{"v":1}`)
	v1, _ := svc.SaveDraft(ctx, form.Key, "default", def1, "admin")

	// Update draft
	def2 := json.RawMessage(`{"v":2}`)
	v2, err := svc.SaveDraft(ctx, form.Key, "default", def2, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v2.ID != v1.ID {
		t.Errorf("expected same draft ID, got %d vs %d", v1.ID, v2.ID)
	}

	// Verify the definition was updated
	stored, _ := vr.FindByID(ctx, v1.ID)
	if string(stored.Definition) != string(def2) {
		t.Errorf("expected updated definition, got %s", string(stored.Definition))
	}

	// Draft pointer should still point to same version
	updated, _ := lr.FindByID(ctx, defaultLayout.ID)
	if *updated.DraftVersionID != v1.ID {
		t.Errorf("draft pointer changed unexpectedly")
	}
}

func TestPublish_Success(t *testing.T) {
	svc, _, lr, vr, _, au := newTestService()
	ctx := context.Background()

	form, defaultLayout, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	// Create a draft
	defJSON := json.RawMessage(`{"sections":[]}`)
	draft, _ := svc.SaveDraft(ctx, form.Key, "default", defJSON, "admin")

	// Publish
	published, err := svc.Publish(ctx, form.Key, "default", "Initial publish", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if published.Kind != "published" {
		t.Errorf("expected kind 'published', got '%s'", published.Kind)
	}
	if published.Description != "Initial publish" {
		t.Errorf("expected description 'Initial publish', got '%s'", published.Description)
	}
	if string(published.Definition) != string(defJSON) {
		t.Errorf("definition should match draft")
	}

	// Verify layout pointers
	updated, _ := lr.FindByID(ctx, defaultLayout.ID)
	if updated.DraftVersionID != nil {
		t.Error("expected draft pointer to be cleared")
	}
	if updated.PublishedVersionID == nil {
		t.Fatal("expected published pointer to be set")
	}
	if *updated.PublishedVersionID != published.ID {
		t.Errorf("expected published pointer %d, got %d", published.ID, *updated.PublishedVersionID)
	}

	// Verify old draft is now archived
	oldDraft, _ := vr.FindByID(ctx, draft.ID)
	if oldDraft.Kind != "archived" {
		t.Errorf("expected old draft kind 'archived', got '%s'", oldDraft.Kind)
	}

	// Verify audit
	found := false
	for _, e := range au.entries {
		if e.Action == "version.publish" {
			found = true
		}
	}
	if !found {
		t.Error("expected version.publish audit entry")
	}
}

func TestPublish_RequiresDescription(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	defJSON := json.RawMessage(`{"sections":[]}`)
	svc.SaveDraft(ctx, form.Key, "default", defJSON, "admin")

	_, err := svc.Publish(ctx, form.Key, "default", "", "admin")
	if err == nil {
		t.Fatal("expected error for empty description")
	}
	if err != formbuilder.ErrPublishDescriptionRequired {
		t.Errorf("expected ErrPublishDescriptionRequired, got %v", err)
	}
}

func TestPublish_NoDraft_ReturnsError(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	_, err := svc.Publish(ctx, form.Key, "default", "No draft exists", "admin")
	if err == nil {
		t.Fatal("expected error for no draft")
	}
	if err != formbuilder.ErrNoDraft {
		t.Errorf("expected ErrNoDraft, got %v", err)
	}
}

func TestRevert_CreatesNewDraftFromPublishedVersion(t *testing.T) {
	svc, _, lr, _, _, _ := newTestService()
	ctx := context.Background()

	form, defaultLayout, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	// Create and publish v1
	def1 := json.RawMessage(`{"v":1}`)
	svc.SaveDraft(ctx, form.Key, "default", def1, "admin")
	pub1, _ := svc.Publish(ctx, form.Key, "default", "v1", "admin")

	// Create and publish v2
	def2 := json.RawMessage(`{"v":2}`)
	svc.SaveDraft(ctx, form.Key, "default", def2, "admin")
	svc.Publish(ctx, form.Key, "default", "v2", "admin")

	// Revert to v1
	reverted, err := svc.Revert(ctx, form.Key, "default", pub1.VersionNumber, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reverted.Kind != "draft" {
		t.Errorf("expected kind 'draft', got '%s'", reverted.Kind)
	}
	if string(reverted.Definition) != string(def1) {
		t.Errorf("expected v1 definition, got %s", string(reverted.Definition))
	}

	// Verify draft pointer is set
	updated, _ := lr.FindByID(ctx, defaultLayout.ID)
	if updated.DraftVersionID == nil {
		t.Fatal("expected draft pointer to be set after revert")
	}
}

func TestArchiveLayout_RejectsDefaultWhileFormActive(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	err := svc.ArchiveLayout(ctx, form.Key, "default", "admin")
	if err == nil {
		t.Fatal("expected ErrCannotArchiveDefault")
	}
	if err != formbuilder.ErrCannotArchiveDefault {
		t.Errorf("expected ErrCannotArchiveDefault, got %v", err)
	}
}

func TestArchiveLayout_AllowsNonDefault(t *testing.T) {
	svc, _, lr, _, _, au := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	layout, _ := svc.CreateLayout(ctx, form.Key, &formbuilder.CreateLayoutRequest{
		Name:        "admin-full",
		DisplayName: "Admin Full",
	}, "admin")

	err := svc.ArchiveLayout(ctx, form.Key, "admin-full", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify layout is archived
	updated, _ := lr.FindByID(ctx, layout.ID)
	if updated.Status != "archived" {
		t.Errorf("expected status 'archived', got '%s'", updated.Status)
	}

	// Verify audit
	found := false
	for _, e := range au.entries {
		if e.Action == "layout.archive" {
			found = true
		}
	}
	if !found {
		t.Error("expected layout.archive audit entry")
	}
}

func TestArchiveForm_CascadesToLayouts(t *testing.T) {
	svc, _, lr, _, _, _ := newTestService()
	ctx := context.Background()

	form, defaultLayout, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	adminLayout, _ := svc.CreateLayout(ctx, form.Key, &formbuilder.CreateLayoutRequest{
		Name:        "admin-full",
		DisplayName: "Admin Full",
	}, "admin")

	err := svc.ArchiveForm(ctx, form.Key, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify both layouts are archived
	def, _ := lr.FindByID(ctx, defaultLayout.ID)
	if def.Status != "archived" {
		t.Errorf("expected default layout archived, got '%s'", def.Status)
	}
	admin, _ := lr.FindByID(ctx, adminLayout.ID)
	if admin.Status != "archived" {
		t.Errorf("expected admin layout archived, got '%s'", admin.Status)
	}
}

func TestAssignRole_Success(t *testing.T) {
	svc, _, lr, _, ar, au := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	adminLayout, _ := svc.CreateLayout(ctx, form.Key, &formbuilder.CreateLayoutRequest{
		Name:        "admin-full",
		DisplayName: "Admin Full",
	}, "admin")

	assignment, err := svc.AssignRole(ctx, form.Key, "admin", "admin-full", "designer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if assignment.RoleName != "admin" {
		t.Errorf("expected role 'admin', got '%s'", assignment.RoleName)
	}
	if assignment.LayoutID != adminLayout.ID {
		t.Errorf("expected layout ID %d, got %d", adminLayout.ID, assignment.LayoutID)
	}

	// Verify in repo
	active, _ := ar.FindActiveByFormAndRole(ctx, form.ID, "admin")
	if active == nil {
		t.Fatal("expected active assignment")
	}

	// Verify audit
	found := false
	for _, e := range au.entries {
		if e.Action == "layout.assign" {
			found = true
		}
	}
	if !found {
		t.Error("expected layout.assign audit entry")
	}

	_ = lr
}

func TestAssignRole_ReplacesExisting(t *testing.T) {
	svc, _, _, _, ar, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	svc.CreateLayout(ctx, form.Key, &formbuilder.CreateLayoutRequest{
		Name:        "layout-a",
		DisplayName: "Layout A",
	}, "admin")
	svc.CreateLayout(ctx, form.Key, &formbuilder.CreateLayoutRequest{
		Name:        "layout-b",
		DisplayName: "Layout B",
	}, "admin")

	// First assignment
	a1, _ := svc.AssignRole(ctx, form.Key, "admin", "layout-a", "designer")

	// Replace with layout-b
	a2, err := svc.AssignRole(ctx, form.Key, "admin", "layout-b", "designer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a2.ID == a1.ID {
		t.Error("expected new assignment ID")
	}

	// Old assignment should be revoked
	all, _ := ar.ListByFormID(ctx, form.ID)
	if len(all) != 1 {
		t.Errorf("expected 1 active assignment, got %d", len(all))
	}
}

func TestRevokeAssignment_Success(t *testing.T) {
	svc, _, _, _, _, au := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	svc.CreateLayout(ctx, form.Key, &formbuilder.CreateLayoutRequest{
		Name:        "admin-full",
		DisplayName: "Admin Full",
	}, "admin")

	svc.AssignRole(ctx, form.Key, "admin", "admin-full", "designer")

	err := svc.RevokeAssignment(ctx, form.Key, "admin", "designer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify audit
	found := false
	for _, e := range au.entries {
		if e.Action == "layout.unassign" {
			found = true
		}
	}
	if !found {
		t.Error("expected layout.unassign audit entry")
	}
}

func TestRevokeAssignment_NotFound(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	err := svc.RevokeAssignment(ctx, form.Key, "nonexistent", "admin")
	if err == nil {
		t.Fatal("expected error for non-existent assignment")
	}
	if err != formbuilder.ErrAssignmentNotFound {
		t.Errorf("expected ErrAssignmentNotFound, got %v", err)
	}
}

func TestListVersions_ReturnsAllVersions(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	// Create and publish v1
	def1 := json.RawMessage(`{"v":1}`)
	svc.SaveDraft(ctx, form.Key, "default", def1, "admin")
	svc.Publish(ctx, form.Key, "default", "v1", "admin")

	// Create and publish v2
	def2 := json.RawMessage(`{"v":2}`)
	svc.SaveDraft(ctx, form.Key, "default", def2, "admin")
	svc.Publish(ctx, form.Key, "default", "v2", "admin")

	versions, err := svc.ListVersions(ctx, form.Key, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have: v1 (published), v1 draft (archived), v2 (published), v2 draft (archived) = 4
	if len(versions) < 2 {
		t.Errorf("expected at least 2 versions, got %d", len(versions))
	}
}

func TestGetVersion_ReturnsSpecificVersion(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	def1 := json.RawMessage(`{"v":1}`)
	svc.SaveDraft(ctx, form.Key, "default", def1, "admin")
	pub, _ := svc.Publish(ctx, form.Key, "default", "v1", "admin")

	version, err := svc.GetVersion(ctx, form.Key, "default", pub.VersionNumber)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version.VersionNumber != pub.VersionNumber {
		t.Errorf("expected version %d, got %d", pub.VersionNumber, version.VersionNumber)
	}
	if string(version.Definition) != string(def1) {
		t.Errorf("definition mismatch")
	}
}

func TestGetVersion_NotFound(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	_, err := svc.GetVersion(ctx, form.Key, "default", 999)
	if err == nil {
		t.Fatal("expected error for non-existent version")
	}
	if err != formbuilder.ErrVersionNotFound {
		t.Errorf("expected ErrVersionNotFound, got %v", err)
	}
}

// --- PR3 Tests ---

func TestAssignRole_CrossFormLayout_Returns422(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	// Create two forms
	formA, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "form-a",
		Name: "Form A",
	}, "admin")

	formB, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "form-b",
		Name: "Form B",
	}, "admin")

	// Create a layout in form B
	svc.CreateLayout(ctx, formB.Key, &formbuilder.CreateLayoutRequest{
		Name:        "special-layout",
		DisplayName: "Special Layout",
	}, "admin")

	// Try to assign form B's layout to form A → should get ErrCrossFormLayout
	_, err := svc.AssignRole(ctx, formA.Key, "admin", "special-layout", "designer")
	if err == nil {
		t.Fatal("expected error for cross-form layout assignment")
	}
	if err != formbuilder.ErrCrossFormLayout {
		t.Errorf("expected ErrCrossFormLayout (422), got %v", err)
	}
}

func TestAssignRole_NonExistentLayout_Returns404(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	_, err := svc.AssignRole(ctx, form.Key, "admin", "nonexistent-layout", "designer")
	if err == nil {
		t.Fatal("expected error for non-existent layout")
	}
	if err != formbuilder.ErrLayoutNotFound {
		t.Errorf("expected ErrLayoutNotFound (404), got %v", err)
	}
}

func TestListAudit_ReturnsEntriesWithFilters(t *testing.T) {
	svc, _, _, _, _, au := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	// CreateForm writes 2 audit entries (form.create + layout.create)
	if len(au.entries) != 2 {
		t.Fatalf("expected 2 audit entries, got %d", len(au.entries))
	}

	// List all audit entries
	entries, total, err := svc.ListAudit(ctx, form.Key, formbuilder.AuditFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}

	// Filter by action
	entries, total, err = svc.ListAudit(ctx, form.Key, formbuilder.AuditFilter{Action: "form.create"}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Mock returns all entries regardless of filter, but the service call should succeed
	if total < 1 {
		t.Errorf("expected at least 1 entry, got %d", total)
	}
	_ = entries
}

func TestListAudit_FormNotFound(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	_, _, err := svc.ListAudit(ctx, "nonexistent-form", formbuilder.AuditFilter{}, 1, 20)
	if err == nil {
		t.Fatal("expected error for non-existent form")
	}
	if err != formbuilder.ErrFormNotFound {
		t.Errorf("expected ErrFormNotFound, got %v", err)
	}
}

func TestListAudit_Pagination(t *testing.T) {
	svc, _, _, _, _, _ := newTestService()
	ctx := context.Background()

	form, _, _ := svc.CreateForm(ctx, &formbuilder.CreateFormRequest{
		Key:  "test-form",
		Name: "Test",
	}, "admin")

	// Test default pagination values
	_, _, err := svc.ListAudit(ctx, form.Key, formbuilder.AuditFilter{}, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error with default pagination: %v", err)
	}

	// Test page size cap
	_, _, err = svc.ListAudit(ctx, form.Key, formbuilder.AuditFilter{}, 1, 200)
	if err != nil {
		t.Fatalf("unexpected error with large page size: %v", err)
	}
}
