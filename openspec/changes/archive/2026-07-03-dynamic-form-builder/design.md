# Design: Dynamic Form Builder

## Technical Approach

Greenfield module following Nova's hexagonal architecture. Backend: `domain/formbuilder` (zero adapter imports) → `adapters/{api,db}/formbuilder` → `infrastructure/cache`. Frontend: `features/form-builder/{designer,runtime}` with signal-based standalone components. Shared Zod schemas in `shared/form-schemas/` consumed by Go (`go:embed`) and Angular (`tsconfig` paths). All tenant-scoped tables live in the tenant schema via `RunInTenantTx` — no `tenant_id` columns. Layouts are role-keyed complete JSON documents; runtime resolution is a join (assignment lookup → published version), not a filter. The `default` layout is auto-created per form and serves as implicit fallback for unassigned roles.

## Architecture Decisions

| # | Decision | Chosen | Rejected | Rationale |
|---|----------|--------|----------|-----------|
| D1 | Tenancy | `search_path` via `RunInTenantTx`, no `tenant_id` columns | Row-level `tenant_id` | Matches conventions §1; isolation is middleware's job |
| D2 | Auth | JWT = identity; active role from `c.Locals("activeRole")` | Role in JWT claim | Conventions §2-§3; role is session state, not identity |
| D3 | Permissions | `HasPermission("formbuilder", action)` with 5 semantic actions | Legacy CRUD columns; per-form permissions | Conventions §5; semantic actions name what they do |
| D4 | Layout model | Role-keyed complete JSON per layout | Role-filtered single definition | Avoids schema drift, muddy reviews, runtime policy walk (pre-proposal §2.1) |
| D5 | Versioning | Per-layout, append-only, immutable published snapshots | Per-form versioning; optimistic locking | Independent evolution per role audience; audit-friendly |
| D6 | `default` layout | Auto-created on form insert, reserved name, service-layer invariants | DB CHECK constraint | CHECK would reject the auto-creation itself; service layer can write audit |
| D7 | Validation | Shared Zod in `shared/form-schemas/`, Go `go:embed` | Native Go structs + TS duplication | Single source of truth; drift is the failure mode we prevent |
| D8 | Cache | In-process `sync.Map`, 10min TTL, key=`{role}:{version_id}` | Redis; pattern-based invalidation | Zero infra for v1; publish bumps version_id → new key; old expires on TTL |
| D9 | Frontend state | Signal-based stores, no NgRx | NgRx; BehaviorSubject | Feature-scoped, signals sufficient, matches codebase pattern |
| D10 | Submissions | Deferred to v2 | Ship in v1 | Keeps scope tight; next feature decides what to do with data |

## Data Flow

### Runtime Resolution (read path)

```
Client GET /api/formbuilder/forms/:formKey
  → ExtractTenant (search_path set)
  → Authenticate (JWT → c.Locals("user"))
  → ContextLoader (ses_active_role → c.Locals("activeRole"))
  → Handler: read role from Locals, check HasPermission("formbuilder","view")
  → Service.Resolve(ctx, formKey, roleName)
      → RunInTenantTx:
          1. SELECT assignment WHERE form+role+revoked_at IS NULL
          2. (if miss) SELECT layout WHERE form+name='default'
          3. SELECT published version by fl_published_version_id
          4. Cache check: key={role}:{version_id}
          5. (if miss) validate JSON, cache write, return
  → Return layout JSON
```

### Publish Flow (write path)

```
Client POST /forms/:formKey/layouts/:layoutName/publish {description}
  → Permission: formbuilder.publish
  → Service.Publish(ctx, formKey, layoutName, description, actor)
      → RunInTenantTx:
          1. Load draft version (fl_draft_version_id)
          2. INSERT new row in eamform_layout_versions (kind='published', version_number=max+1)
          3. UPDATE eamform_layouts SET fl_published_version_id = new_id
          4. INSERT audit row (action='version.publish', metadata={from,to,description})
  → Cache: old key expires on TTL; new version → new key on next read
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `backend/migrations/tenant/20260219000001_form_definitions.up.sql` | Create | Tables, indexes, triggers, constraints for eamform_* |
| `backend/migrations/tenant/20260219000001_form_definitions.down.sql` | Create | Drop eamform_* tables |
| `backend/migrations/tenant/20260219000002_form_submissions.up.sql` | Create | eamform_submissions (v2 placeholder) |
| `backend/migrations/tenant/20260219000003_seed_form_designer_role.up.sql` | Create | Seed form_designer role + permissions |
| `backend/internal/domain/formbuilder/entity.go` | Create | Form, Layout, LayoutVersion, Assignment, AuditEntry structs |
| `backend/internal/domain/formbuilder/ports.go` | Create | Repository interfaces (Form, Layout, Version, Assignment, Audit) |
| `backend/internal/domain/formbuilder/service.go` | Create | Business logic + default-layout invariants |
| `backend/internal/domain/formbuilder/resolve.go` | Create | Resolution algorithm (assignment → default fallback → version) |
| `backend/internal/domain/formbuilder/dto.go` | Create | HTTP request/response DTOs |
| `backend/internal/domain/formbuilder/errors.go` | Create | Sentinel errors (ReservedLayoutName, CannotArchiveDefault, etc.) |
| `backend/internal/adapters/api/formbuilder/handler.go` | Create | Fiber handlers for all endpoints |
| `backend/internal/adapters/api/formbuilder/routes.go` | Create | Route registration under /api/formbuilder |
| `backend/internal/adapters/api/formbuilder/middleware.go` | Create | Permission gate helper |
| `backend/internal/adapters/db/formbuilder/form_repository.go` | Create | PgX implementation of FormRepository |
| `backend/internal/adapters/db/formbuilder/layout_repository.go` | Create | PgX implementation of LayoutRepository |
| `backend/internal/adapters/db/formbuilder/layout_version_repository.go` | Create | PgX implementation of LayoutVersionRepository |
| `backend/internal/adapters/db/formbuilder/assignment_repository.go` | Create | PgX implementation of AssignmentRepository |
| `backend/internal/adapters/db/formbuilder/audit_log_repository.go` | Create | PgX implementation of AuditLogRepository |
| `backend/internal/infrastructure/cache/layout_cache.go` | Create | sync.Map cache with TTL, LayoutCache interface |
| `backend/internal/infrastructure/wire/wire.go` | Modify | Wire formbuilder repos → service → handler |
| `shared/form-schemas/layout-definition.schema.ts` | Create | Zod schema for layout JSON |
| `shared/form-schemas/field-types.ts` | Create | Field type discriminated union |
| `shared/form-schemas/validators.ts` | Create | Validator kind catalog |
| `shared/form-schemas/rules.ts` | Create | Cross-field rule catalog |
| `shared/form-schemas/index.ts` | Create | Public re-exports |
| `frontend/src/app/features/form-builder/form-builder.routes.ts` | Create | Lazy-loaded route config |
| `frontend/src/app/features/form-builder/runtime/` | Create | FormRuntimeComponent + 8 field renderers + cross-field validator |
| `frontend/src/app/features/form-builder/designer/` | Create | FormDesignerComponent + CDK DragDrop canvas + palette + settings |
| `frontend/src/app/features/form-builder/state/` | Create | Signal-based stores (designer + runtime) |
| `frontend/src/app/features/form-builder/models/` | Create | TS types mirroring shared schemas |
| `frontend/src/app/features/form-builder/services/` | Create | HTTP clients (designer, runtime, assignment) |
| `frontend/src/app/app.routes.ts` | Modify | Add lazy route for form-builder feature |

## Interfaces / Contracts

### API Endpoints (v1 — submissions deferred)

| Method | Path | Permission | Body |
|--------|------|------------|------|
| GET | `/api/formbuilder/forms` | `view` | — |
| POST | `/api/formbuilder/forms` | `design` | `{key, name, description}` |
| POST | `/api/formbuilder/forms/:formKey/archive` | `publish` | — |
| GET | `/api/formbuilder/forms/:formKey` | `view` | — (runtime resolve) |
| GET | `/api/formbuilder/forms/:formKey/layouts` | `view` | — |
| POST | `/api/formbuilder/forms/:formKey/layouts` | `design` | `{name, displayName, description}` |
| POST | `/api/formbuilder/forms/:formKey/layouts/:layoutName/archive` | `publish` | — |
| GET | `/api/formbuilder/forms/:formKey/layouts/:layoutName/draft` | `view_draft` | — |
| PUT | `/api/formbuilder/forms/:formKey/layouts/:layoutName/draft` | `design` | layout JSON |
| POST | `/api/formbuilder/forms/:formKey/layouts/:layoutName/publish` | `publish` | `{description}` |
| POST | `/api/formbuilder/forms/:formKey/layouts/:layoutName/revert` | `design` | `{versionNumber}` |
| GET | `/api/formbuilder/forms/:formKey/layouts/:layoutName/versions` | `view` | — |
| GET | `/api/formbuilder/forms/:formKey/layouts/:layoutName/versions/:n` | `view` | — |
| GET | `/api/formbuilder/forms/:formKey/assignments` | `view` | — |
| PUT | `/api/formbuilder/forms/:formKey/assignments/:roleName` | `assign` | `{layoutName}` |
| DELETE | `/api/formbuilder/forms/:formKey/assignments/:roleName` | `assign` | — |
| GET | `/api/formbuilder/forms/:formKey/audit` | `view` | query: page, action, entity_type |

### Domain Entities (Go)

```go
// domain/formbuilder/entity.go — no tenant_id fields
type Form struct {
    ID          int64     `json:"frm_id"`
    Key         string    `json:"frm_key"`
    Name        string    `json:"frm_name"`
    Description string    `json:"frm_description"`
    Status      string    `json:"frm_status"` // active|archived
    CreatedBy   string    `json:"frm_created_by"`
    CreatedAt   time.Time `json:"frm_created_at"`
    UpdatedAt   time.Time `json:"frm_updated_at"`
}

type Layout struct {
    ID                int64   `json:"fl_id"`
    FormID            int64   `json:"fl_form_id"`
    Name              string  `json:"fl_name"`
    DisplayName       string  `json:"fl_display_name"`
    Description       string  `json:"fl_description"`
    Status            string  `json:"fl_status"`
    DraftVersionID    *int64  `json:"fl_draft_version_id"`
    PublishedVersionID *int64 `json:"fl_published_version_id"`
    CreatedBy         string  `json:"fl_created_by"`
    CreatedAt         time.Time `json:"fl_created_at"`
    UpdatedAt         time.Time `json:"fl_updated_at"`
}

type LayoutVersion struct {
    ID            int64           `json:"flv_id"`
    LayoutID      int64           `json:"flv_layout_id"`
    VersionNumber int             `json:"flv_version_number"`
    Kind          string          `json:"flv_kind"` // draft|published|archived
    Description   string          `json:"flv_description"`
    Definition    json.RawMessage `json:"flv_definition"`
    CreatedBy     string          `json:"flv_created_by"`
    CreatedAt     time.Time       `json:"flv_created_at"`
    PublishedAt   *time.Time      `json:"flv_published_at"`
}
```

### Resolution Signature

```go
// domain/formbuilder/service.go
func (s *FormService) Resolve(ctx context.Context, formKey string, roleName string) (*LayoutVersion, error)
// No tenantID parameter — RunInTenantTx already scoped the connection.
// Returns FormDefaultLayoutMissing (500) only if default layout row is absent (data-integrity bug).
```

### DB Schema (key constraints)

```sql
-- Partial unique: one draft per layout
CREATE UNIQUE INDEX uq_flv_one_draft ON eamform_layout_versions (flv_layout_id)
    WHERE flv_kind = 'draft';

-- Partial unique: one active assignment per (form, role)
CREATE UNIQUE INDEX uq_fra_active ON eamform_role_assignments (fra_form_id, fra_role_name)
    WHERE fra_revoked_at IS NULL;

-- Immutability triggers
CREATE TRIGGER trg_no_update_versions BEFORE UPDATE ON eamform_layout_versions
    FOR EACH STATEMENT EXECUTE FUNCTION raise_immutable();
CREATE TRIGGER trg_no_update_audit BEFORE UPDATE ON eamform_audit_log
    FOR EACH STATEMENT EXECUTE FUNCTION raise_immutable();
CREATE TRIGGER trg_no_delete_audit BEFORE DELETE ON eamform_audit_log
    FOR EACH STATEMENT EXECUTE FUNCTION raise_immutable();

-- Protect default layout from orphaned form deletion
CREATE TRIGGER trg_protect_default_layout BEFORE DELETE ON eamform_definitions
    FOR EACH ROW EXECUTE FUNCTION check_default_layout_exists();
```

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | Service invariants (default-layout rules, version numbering) | Go table-driven tests with mock repos |
| Integration | Resolve algorithm, publish flow, audit writes | `go test` with test DB, `RunInTenantTx` against real pgxmock |
| Integration | Immutability triggers | Direct SQL UPDATE/DELETE assertions |
| Frontend unit | Field renderers, cross-field validator, store logic | Jasmine + MSW fixtures |
| Frontend integration | Designer drag-reorder serialization, publish flow | MSW-based component tests |
| E2E (PR6) | Design → publish → assign → resolve → render round-trip | Full stack against seeded DB |

## Migration / Rollout

**Prerequisites (must land before PR1):**
- Permissions migration: `eamrole_permissions` must support `(screen, action, allowed)` rows — the legacy CRUD columns (`rpe_select`, `rpe_insert`, etc.) block `HasPermission("formbuilder", "design")` from working. Verify `Role.HasPermission` reads from the normalized model; if not, migrate first.

**Phased delivery (6 chained PRs):**
1. **PR1** — Backend foundation: migrations, domain, read endpoints, default-layout invariants, cache, seed. ~600 LOC.
2. **PR2** — Backend design+publish: draft save, publish, revert, archive, assignments. ~550 LOC.
3. **PR3** — Backend audit: audit endpoint, immutability tests. ~200 LOC.
4. **PR4** — Frontend runtime: renderer, field components, reactive forms, shared schemas. ~700 LOC.
5. **PR5** — Frontend designer: visual builder, CDK DragDrop, layout picker, assignment panel. ~950 LOC.
6. **PR6** — Integration: cross-field rules, submissions endpoint, e2e tests. ~400 LOC.

**Rollback per PR:** Each migration has a `.down.sql`; route group removal reverts API surface; `form_designer` seed removal reverts permissions.

## Open Questions

- [ ] **Permissions prerequisite:** Does the current `eamrole_permissions` table already support the normalized `(screen, action, allowed)` model, or does a migration from legacy CRUD columns need to land first? The `Role.HasPermission` method exists and works on `map[string]map[string]bool`, but the DB loading code must populate it from the normalized schema. This is the single blocker for PR1.
