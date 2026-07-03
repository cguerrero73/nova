# Tasks: Dynamic Form Builder

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~3,350 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR0 → PR1 → PR2 → PR3 → PR4 → PR5 → PR6 |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 0 | Permissions migration (CRUD→normalized) | PR0 | **Blocker** — PR1 cannot land without it. ~150 LOC |
| 1 | Backend foundation: schema, domain, read endpoint, cache, seed | PR1 | ~600 LOC. Base branch: main |
| 2 | Backend design+publish: draft, publish, revert, archive, assignments | PR2 | ~550 LOC. Base: PR1 branch |
| 3 | Backend audit log endpoint + immutability tests | PR3 | ~200 LOC. Base: PR2 branch |
| 4 | Frontend runtime renderer + shared Zod schemas | PR4 | ~700 LOC. Base: main (independent of backend PRs for dev) |
| 5 | Frontend visual designer + CDK DragDrop | PR5 | ~950 LOC. Base: PR4 branch |
| 6 | Integration: cross-field rules, e2e, polish | PR6 | ~400 LOC. Base: PR5 branch |

## PR0: Permissions Prerequisite

- [x] 0.1 Create migration `20260218000000_normalize_role_permissions.up.sql`: add `rpe_action VARCHAR(50)`, `rpe_allowed BOOLEAN` columns; migrate existing CRUD rows into normalized rows (one per action); drop legacy columns. `backend/migrations/tenant/`
- [x] 0.2 Create matching `.down.sql` rollback.
- [x] 0.3 Update role repository query to load `eamrole_permissions` into `map[string]map[string]bool` using `(rpe_screen, rpe_action, rpe_allowed)`. `backend/internal/adapters/db/roles/`
- [x] 0.4 Verify: `go test ./...` passes; `HasPermission("formbuilder","design")` works with seeded data.

**Dependencies**: None. **LOC**: ~150. **Blocks**: PR1.

## PR1: Backend Foundation

- [x] 1.1 Create migration `20260219000001_form_definitions.up.sql`: `eamform_definitions`, `eamform_layouts`, `eamform_layout_versions`, `eamform_role_assignments` tables, indexes, immutability triggers, partial uniques. `backend/migrations/tenant/`
- [x] 1.2 Create matching `.down.sql`.
- [x] 1.3 Create seed migration `20260219000003_seed_form_designer_role.up.sql`: insert `form_designer` role + 5 permission rows.
- [x] 1.4 Create `domain/formbuilder/entity.go`: Form, Layout, LayoutVersion, Assignment structs (no tenant_id).
- [x] 1.5 Create `domain/formbuilder/errors.go`: sentinel errors (ReservedLayoutName, CannotArchiveDefault, FormLayoutNotPublished, FormDefaultLayoutMissing).
- [x] 1.6 Create `domain/formbuilder/ports.go`: repository interfaces (FormRepository, LayoutRepository, LayoutVersionRepository, AssignmentRepository).
- [x] 1.7 Create `domain/formbuilder/service.go`: CreateForm (with default layout auto-create in same tx), GetForm, ListForms, ArchiveForm.
- [x] 1.8 Create `domain/formbuilder/resolve.go`: Resolve(ctx, formKey, roleName) — assignment lookup → default fallback → published version.
- [x] 1.9 Create `infrastructure/cache/layout_cache.go`: sync.Map with 10min TTL, key=`{role}:{version_id}`.
- [x] 1.10 Create `adapters/db/formbuilder/form_repository.go`, `layout_repository.go`, `layout_version_repository.go`, `assignment_repository.go`: PgX implementations.
- [x] 1.11 Create `adapters/api/formbuilder/handler.go`: GET /forms, GET /forms/:formKey, GET /forms/:formKey/layouts.
- [x] 1.12 Create `adapters/api/formbuilder/routes.go` + `middleware.go`: route registration under /api/formbuilder, permission gate helper.
- [x] 1.13 Modify `infrastructure/wire/wire.go`: add formbuilder repos → service → handler to Container.
- [x] 1.14 Tests: table-driven service tests (mock repos), integration test for Resolve with pgxmock.

**Dependencies**: PR0. **LOC**: ~600.

## PR2: Backend Design + Publish

- [ ] 2.1 Add `CreateLayout` to service: reject reserved name `default`, reject duplicates, write audit. `domain/formbuilder/service.go`
- [ ] 2.2 Add `SaveDraft`: create or update draft version row, enforce one-draft-per-layout. `domain/formbuilder/service.go`
- [ ] 2.3 Add `Publish`: create published version from draft, bump version_number, clear draft pointer, write audit. `domain/formbuilder/service.go`
- [ ] 2.4 Add `Revert`: copy published version N back as new draft. `domain/formbuilder/service.go`
- [ ] 2.5 Add `ArchiveLayout`: soft-archive, reject default while form active. `domain/formbuilder/service.go`
- [ ] 2.6 Add assignment endpoints to service: AssignRole, RevokeAssignment, ListAssignments. `domain/formbuilder/service.go`
- [ ] 2.7 Create `domain/formbuilder/dto.go`: HTTP request/response DTOs.
- [ ] 2.8 Extend `adapters/api/formbuilder/handler.go`: POST forms, PUT/POST draft, POST publish, POST revert, POST archive (form+layout), PUT/DELETE assignments, GET versions, GET assignments.
- [ ] 2.9 Extend `adapters/db/formbuilder/`: audit_log_repository.go PgX implementation.
- [ ] 2.10 Tests: publish flow integration, draft save/update, assignment replace+revoke, archive invariants.

**Dependencies**: PR1. **LOC**: ~550.

## PR3: Backend Audit

- [ ] 3.1 Create migration `20260219000002_form_audit.up.sql`: `eamform_audit_log` table + immutability triggers (no UPDATE, no DELETE).
- [ ] 3.2 Add `AuditLogRepository` interface to `domain/formbuilder/ports.go`.
- [ ] 3.3 Wire audit writes into all service mutations (form create/archive, layout create/archive, version publish/revert, assignment assign/revoke).
- [ ] 3.4 Add GET /forms/:formKey/audit handler with pagination + filters (action, entity_type, date range).
- [ ] 3.5 Tests: immutability triggers (direct SQL UPDATE/DELETE assertions), audit retrieval with filters, actor attribution.

**Dependencies**: PR2. **LOC**: ~200.

## PR4: Frontend Runtime Renderer

- [ ] 4.1 Create `shared/form-schemas/layout_definition.schema.ts`: Zod schema for layout JSON.
- [ ] 4.2 Create `shared/form-schemas/field-types.ts`: discriminated union for 8 field types.
- [ ] 4.3 Create `shared/form-schemas/validators.ts`: validator kind catalog.
- [ ] 4.4 Create `shared/form-schemas/rules.ts`: cross-field rule catalog (equals, notEquals, requiredIf, hiddenIf).
- [ ] 4.5 Create `shared/form-schemas/index.ts`: public re-exports.
- [ ] 4.6 Create `frontend/src/app/features/form-builder/models/`: TS types mirroring shared schemas.
- [ ] 4.7 Create `frontend/src/app/features/form-builder/services/runtime.service.ts`: HTTP client for GET /forms/:formKey.
- [ ] 4.8 Create `frontend/src/app/features/form-builder/runtime/form-runtime.component.ts`: signal-based standalone, Reactive Forms, section/field rendering.
- [ ] 4.9 Create 8 field renderer components (text, textarea, number, date, checkbox, select, radio, multiselect) under `runtime/fields/`.
- [ ] 4.10 Create `runtime/cross-field-validator.ts`: evaluates rules[] against FormGroup values.
- [ ] 4.11 Create `form-builder.routes.ts`: lazy route for runtime; modify `app.routes.ts`.
- [ ] 4.12 Tests: field renderers, grid layout (half/third), required+pattern validation, requiredIf+equals rules.

**Dependencies**: None (parallel with backend). **LOC**: ~700.

## PR5: Frontend Designer

- [ ] 5.1 Create `frontend/src/app/features/form-builder/services/designer.service.ts`: HTTP client for CRUD forms/layouts/drafts/publish/versions.
- [ ] 5.2 Create `frontend/src/app/features/form-builder/services/assignment.service.ts`: HTTP client for assignments CRUD.
- [ ] 5.3 Create `frontend/src/app/features/form-builder/state/designer.store.ts`: signal-based store for designer state.
- [ ] 5.4 Create `designer/form-designer.component.ts`: three-panel layout (palette, canvas, settings).
- [ ] 5.5 Create `designer/field-palette.component.ts`: 8 draggable field types.
- [ ] 5.6 Create `designer/section-canvas.component.ts`: CDK DragDrop sections + fields, reorder, cross-section move.
- [ ] 5.7 Create `designer/field-settings.component.ts`: reactive config panel (label, placeholder, validators, options for select/radio/multiselect).
- [ ] 5.8 Create `designer/layout-picker.component.ts`: list layouts, default badge, create new.
- [ ] 5.9 Create `designer/assignment-panel.component.ts`: role-to-layout mapping, unassigned roles show "→ uses default".
- [ ] 5.10 Create `designer/preview-dialog.component.ts`: render draft using FormRuntimeComponent.
- [ ] 5.11 Add shared Zod validation before save draft (block + inline errors).
- [ ] 5.12 Tests: drag-reorder serialization, publish flow, assignment panel, shared schema validation.

**Dependencies**: PR4. **LOC**: ~950.

## PR6: Integration + Polish

- [ ] 6.1 Implement cross-field rule evaluation in renderer: hiddenIf visibility toggle, equals/notEquals error messages.
- [ ] 6.2 Create audit panel component in designer: displays audit entries with filters.
- [ ] 6.3 Wire designer audit panel to GET /forms/:formKey/audit.
- [ ] 6.4 E2E test: design → publish → assign → resolve → render round-trip against seeded DB.
- [ ] 6.5 Verify: `go build`, `golangci-lint`, `pnpm build`, `ng test` all pass.
- [ ] 6.6 Verify: no `tenant_id` in any `eamform_*` table; `search_path` isolation confirmed.

**Dependencies**: PR3, PR5. **LOC**: ~400.
