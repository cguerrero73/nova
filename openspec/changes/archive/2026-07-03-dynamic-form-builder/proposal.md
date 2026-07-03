# Proposal: Dynamic Form Builder

## Intent

Nova lacks configurable data capture — forms are hardcoded, requiring deploys for changes. This module adds versioned, role-aware form definitions with a visual designer and runtime renderer, scoped per tenant via schema-per-tenant isolation.

## Scope

**In**: Form definitions, per-layout versioning, role-to-layout assignment with `default` fallback, visual designer (CDK DragDrop), runtime renderer (Reactive Forms), shared Zod validation, audit log, 8 field types, cross-field validators, 5 permission actions.

**Out**: Submissions (v2), dynamic select, multi-locale, audit retention config, per-form permissions, file/richtext/repeater fields, Redis cache.

## Capabilities

**New**: `form-definition`, `form-layout`, `form-role-assignment`, `form-runtime-resolution`, `form-designer`, `form-renderer`, `form-audit`

**Modified**: None (greenfield)

## Approach

- **Tenancy**: Schema-per-tenant via `search_path` (conventions §1). No `tenant_id` columns.
- **Auth**: JWT = identity (conventions §2). Active role from `c.Locals("activeRole")` (conventions §3).
- **Permissions**: `HasPermission(screen, action)` with semantic actions (conventions §5). New `form_designer` role.
- **Versioning**: Per-layout, independent draft + published chains, immutable snapshots on publish.
- **Validation**: Shared Zod in `shared/form-schemas/` (Go `go:embed`, Angular `tsconfig`).
- **Cache**: In-process `sync.Map`, 10min TTL, keyed by `{role}:{version}`.

## Affected Areas

| Area | Impact |
|------|--------|
| `backend/internal/domain/formbuilder/` | New |
| `backend/internal/adapters/{api,db}/formbuilder/` | New |
| `backend/internal/infrastructure/cache/` | New |
| `backend/migrations/tenant/` | 3 files |
| `frontend/src/app/features/form-builder/` | New |
| `shared/form-schemas/` | New |

## Open Questions — Resolved

1. `default` → starter JSON
2. New roles → unassigned, fallback
3. Independent drafts per layout
4. Submissions → v2
5. 5-action permissions
6. TTL 10min cache
7. Static-only select v1
8. CDK DragDrop + drop zones
9. Single-locale v1
10. Audit panel in designer v1
11. `UNIQUE(frm_key)`

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Legacy CRUD permissions block `HasPermission` | Med | Verify prerequisite; migrate if needed |
| ~3,350 LOC exceeds 400-line budget | High | 6 chained PRs mandatory |
| Shared Zod drift Go/Angular | Low | `go:embed` + build-time validation |
| `default` invariants service-layer only | Low | Integration tests |

## Rollback

- Migration `down` drops `eamform_*` tables
- Remove `formbuilder` route group + frontend route
- Remove `form_designer` seed + permissions

## Dependencies

- **Prerequisite**: `Role.HasPermission(screen, action)` works (conventions §5 flags CRUD→generic migration as pending debt)
- **Prerequisite**: `RunInTenantTx` operational (in place)
- **Prerequisite**: Angular CDK available (used by `query-builder`)

## Success Criteria

- [ ] Designer creates forms, layouts, publishes, assigns roles
- [ ] Runtime returns correct layout per role (or `default`)
- [ ] Publish creates immutable snapshots; prior versions queryable
- [ ] All mutations write audit log; panel displays entries
- [ ] Shared Zod validates on Go + Angular
- [ ] No `tenant_id` in `eamform_*` tables
- [ ] `go build`, `golangci-lint`, `pnpm build`, `ng test` pass
