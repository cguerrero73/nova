# Proposal: Fix Tenant Context Propagation Across API Handlers

## Intent

The Fiber tenant middleware stores the tenant in a wrapped Go context via `c.SetUserContext(ctx)` (`backend/internal/infrastructure/middleware/tenant.go:57`). `RunInTenantTx` (`backend/internal/infrastructure/db/context.go:39`) reads it from that wrapper to set the per-tenant `search_path`. Eleven of twelve API handler files call `c.Context()` instead of `c.UserContext()`, which returns the raw `*fasthttp.RequestCtx` — the wrapped valueCtx is lost, `extractTenantCode` returns empty, the transaction runs against `public`, and every per-tenant SELECT/INSERT/UPDATE/DELETE fails with `SQLSTATE 42P01` ("relation does not exist"). Commit `1dc091a` fixed only `grid/handler.go`; the other 11 files still ship the bug.

**User impact:** all non-grid endpoints in those 11 domains (queries, users, stores, events, objects, parts, stocks, syscodes, organizations, structure, auth-adjacent) return 500 errors on any tenant-scoped call. Writes are equally broken (loud 42P01, but still 500 to the user). The grid path is the only one working today.

## Scope

### In Scope
- **(a)** Mechanical replacement: `c.Context()` → `c.UserContext()` at all 65 call sites in 11 handler files:
  - `backend/internal/adapters/api/queries/handler.go` (6)
  - `backend/internal/adapters/api/users/handler.go` (5)
  - `backend/internal/adapters/api/auth/handler.go` (5)
  - `backend/internal/adapters/api/stores/handler.go` (9)
  - `backend/internal/adapters/api/events/handler.go` (7)
  - `backend/internal/adapters/api/objects/handler.go` (6)
  - `backend/internal/adapters/api/parts/handler.go` (5)
  - `backend/internal/adapters/api/stocks/handler.go` (11)
  - `backend/internal/adapters/api/syscodes/handler.go` (6)
  - `backend/internal/adapters/api/organizations/handler.go` (5)
  - `backend/internal/adapters/api/structure/handler.go` (5)
- **(c)** Regression test (`backend/internal/infrastructure/db/context_test.go`): unit test that wraps a context with `TenantContextKey{}` via a fake Fiber ctx and asserts `RunInTenantTx` sees the tenant and emits the `SET search_path` statement. Mock the pool with `pgxmock` (or equivalent) so no live DB is required.

### Out of Scope
- **(b)** Registering the missing `GET /api/v1/queries/:id` route — separate change.
- Removing the fasthttp fallback in `extractTenantCode` — defensive code, will become inert but not deleted in this change.
- Any refactor of the tenant middleware, service layer, or repositories.

## Capabilities

### New Capabilities
None — this is a bug fix, not a new feature.

### Modified Capabilities
None — `openspec/specs/` does not exist in this project (no existing capability covers tenant propagation). Adding a spec for tenant propagation is out of scope for this bug-fix change and should be tracked separately if desired.

## Approach

Single-purpose mechanical change. Each handler file gets a global replace of `c.Context()` → `c.UserContext()`. Both methods return `context.Context`-compatible values and have identical use at call sites, so the change is safe and the surrounding cancellation/timeout semantics are unchanged. The regression test pins the contract so any future handler that reverts to `c.Context()` fails the build.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| 11 handler files under `backend/internal/adapters/api/*/handler.go` | Modified | `c.Context()` → `c.UserContext()` at 65 sites |
| `backend/internal/infrastructure/db/context_test.go` | New | Regression test for tenant propagation into `RunInTenantTx` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| One of 65 sites depends on raw fasthttp ctx behavior (e.g. direct `*fasthttp.RequestCtx` type assertion) | Low | Grep first; if any site asserts `*fasthttp.RequestCtx` explicitly, change the assertion to accept the wrapped ctx. Test suite catches regressions. |
| Reviewer fatigue over 65 mechanical changes | Med | Single PR, one commit, all changes are byte-identical in shape; rely on `git diff` and the regression test rather than per-line review. |
| Hidden ctx contract (deadline, cancellation) shifts between the two accessors | Low | Both are request-scoped; `c.UserContext()` is the documented contract for downstream Go context consumers. |
| Test infra absent in repo (`openspec/config.yaml` reports no tests) | Med | Add a self-contained `go test` file using `pgxmock`; no DB or migration step required. |

## Rollback Plan

Revert the single commit. Because the change is mechanical and the regression test is additive, a `git revert` cleanly restores all 11 handler files to their pre-fix state and removes the test.

## Success Criteria

- [ ] All 65 call sites replaced; `grep -n "c.Context()" backend/internal/adapters/api` returns no matches.
- [ ] New regression test passes: `go test ./internal/infrastructure/db/...`.
- [ ] Manual smoke: `GET /api/v1/queries?gridId=1` (or any previously failing route) against tenant `acme` returns 200 with rows, not 42P01.
- [ ] `go build ./...` passes; `golangci-lint run` passes on changed files.
