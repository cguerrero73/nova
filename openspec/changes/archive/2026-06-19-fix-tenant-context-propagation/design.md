# Design: Fix Tenant Context Propagation Across API Handlers

## Technical Approach

Mechanical, single-line replacements in 11 Fiber handlers: `c.Context()` → `c.UserContext()`. Each makes the wrapped Go context (set by tenant middleware via `c.SetUserContext`) visible to `RunInTenantTx` (`context.go:39`). Regression test pins the contract via `pgxmock`.

## Architecture Decisions

| Decision | Choice | Why |
|----------|--------|-----|
| PR shape | One PR, one commit, no delta spec | Byte-identical replacements at 70 sites — diffstat + test > per-line review. User-approved `sdd-spec` skip (`openspec/specs/` absent). |
| Test path | `backend/internal/infrastructure/db/context_test.go` | `RunInTenantTx` lives in `db/`; no `tenant/` subdir. Prompt's `infrastructure/tenant/runner_test.go` is wrong. |
| Test mock | `github.com/pashagolub/pgxmock/v3` | Captures SQL via `tx.Exec`. v3 = pgx/v5-compatible. No live DB. |
| Test Fiber | Don't fake `*fiber.Ctx` | `RunInTenantTx` only reads its `ctx` param. Pass `WithValue(ctx, TenantContextKey{}, "acme")` directly. |
| Fasthttp fallback | Keep in `extractTenantCode` (179–184) | Inert after fix; removal is separate cleanup. |

## Data Flow

```
HTTP → TenantMiddleware (middleware/tenant.go:28)
       c.SetUserContext(WithValue(c.Context(), db.TenantContextKey{}, "acme"))
     → Handler (queries/handler.go:41) ctx := c.UserContext()  ← was c.Context()
     → Service → Repository (ctx unchanged)
     → RunInTenantTx (context.go:39) extractTenantCode(ctx) → "acme"
       BEGIN; SET search_path TO tenant_acme, public; fn(tx); COMMIT
     → PostgreSQL
```

## File Changes

| File | Sites | File | Sites |
|------|------|------|------|
| `api/auth/handler.go` | 5 (32,59,79,99,118) | `api/parts/handler.go` | 5 (32,42,64,83,97) |
| `api/events/handler.go` | 7 (32,42,64,83,97,108,126) | `api/queries/handler.go` | 6 (41,43,65,94,118,137) |
| `api/objects/handler.go` | 6 (32,42,64,83,97,112) | `api/stocks/handler.go` | 11 (29,43,53,75,94,108,126,148,167,186,200) |
| `api/organizations/handler.go` | 5 (26,36,58,77,91) | `api/stores/handler.go` | 9 (27,37,59,78,92,102,121,140,154) |
| `api/structure/handler.go` | 5 (26,36,58,77,91) | `api/syscodes/handler.go` | 6 (20,30,47,66,80,93) |
| `api/users/handler.go` | 5 (31,41,63,82,96) | | |

Plus: `infrastructure/db/context_test.go` (new), `go.mod`/`go.sum` (add `pgxmock/v3`). **Total: 70 mechanical replacements across 11 files + 1 new test file.** *(Proposal said 65 — see Risks R1.)* All paths are relative to `backend/`.

## Representative Before/After

`backend/internal/adapters/api/queries/handler.go` line 41:

```go
// BEFORE
result, err = h.queryService.List(c.Context(), gridID, userCode)
// AFTER
result, err = h.queryService.List(c.UserContext(), gridID, userCode)
```

Same shape at every site; both return `context.Context`.

## What is NOT a `c.Context()` call (out of scope)

| Pattern | Where | Why excluded |
|---------|-------|--------------|
| `context.Background()` | `context.go:131,137`, `tenant.go:24,49,56`, migrate/main.go | Package-qualified, not Fiber; replacement breaks cancellation. |
| `context.WithValue(...)` | `middleware/tenant.go:55` | Builds the wrapped ctx handed to Fiber. Must stay. |
| `"context."` string literals | None | Audited. |
| `*fasthttp.RequestCtx` assertions | None across 11 files | Audited — zero hits. |

**Reviewer guard**: post-apply `grep -rn 'c\.Context()' backend/internal/adapters/api/` must be empty.

## Regression Test Contract

`backend/internal/infrastructure/db/context_test.go` (`package db`). Table-driven, `pgxmock.QueryMatcherEqual` for byte-exact SQL. Three subtests assert on **captured SQL order/contents** (observable contract), not internal call counts:

| Subtest | ctx | Expected SQL (in order) |
|---------|-----|--------------------------|
| wrapped ctx emits SET search_path before any table query | `WithValue(Background(), TenantContextKey{}, "acme")` | `SET search_path TO tenant_acme, public`, then `SELECT 1` |
| no tenant falls back to public schema only | `context.Background()` | `SELECT 1` only |
| raw fasthttp.RequestCtx with UserValue('tenant') hits legacy fallback | `fctxWithUserValue("legacy")` | `SET search_path TO tenant_legacy, public`, then `SELECT 1` |

Per subtest: `pgxmock.NewPgxPool(QueryMatcherEqual)` → `ExpectBegin` → chained `ExpectExec` per row → call `RunInTenantTx(ctx, mock, fnReturningSelectOne)` → `require.NoError(t, mock.ExpectationsWereMet())`. Third subtest exercises `extractTenantCode`'s fasthttp fallback so any future change is intentional.

## Testing Strategy

| Layer | What | How |
|-------|------|-----|
| Unit | `RunInTenantTx` propagates `TenantContextKey{}` → `SET search_path` | `context_test.go` — pgxmock, 3 subtests above |
| Integration / E2E | None new | Live DB needs testcontainers; deferred. Manual smoke in §Verification. |

## Verification

```bash
# 1. Static guard — 0 stragglers
grep -rn 'c\.Context()' backend/internal/adapters/api/
# 2. Compile
cd backend && go build ./...
# 3. Regression test
cd backend && go test ./internal/infrastructure/db/... -v -run TestRunInTenantTx_TenantPropagation
# 4. Manual smoke
curl -H "X-Tenant-Code: acme" 'http://localhost:8080/api/v1/queries?gridId=1'
# Expect: 200 + rows; log shows `[TX] [acme] BEGIN + SET search_path TO tenant_acme, public`
# 5. Lint
cd backend && golangci-lint run ./internal/adapters/api/... ./internal/infrastructure/db/...
```

## Migration / Rollout

No data migration, no feature flag. Single PR + revert. Fasthttp fallback retained as defensive code; cleanup deferred.

## Risks

1. **Call-site count drift (65 → 70).** Actual `grep -c` (excluding `grid/handler.go:42` comment) = **70**. sdd-tasks must use the 70-site count.
2. **Test infra absent.** Zero `*_test.go` today; new test introduces pgxmock and the first `go test`. CI wiring out of scope.
3. **pgxmock version pin.** Must be `v3` for pgx/v5. sdd-apply: `go get github.com/pashagolub/pgxmock/v3` — never the unversioned path (v2/pgx/v4).
4. **Hidden fasthttp ctx dep at any of 70 sites.** Zero `*fasthttp.RequestCtx` assertions across the 11 files. A missed assertion fails `go build` (return-type mismatch); build step is the safety net.
5. **Fasthttp fallback becomes dead but kept.** Third subtest locks its behavior so future changes are intentional.

## Relevant Files

- `backend/internal/infrastructure/db/context.go` — `RunInTenantTx` (39), `extractTenantCode` (171). Test target.
- `backend/internal/infrastructure/middleware/tenant.go` — wraps ctx at 55–57. NOT modified.
- `backend/internal/adapters/api/grid/handler.go` — canonical `c.UserContext()` (43, 81, 114). NOT modified.
- 11 handlers in §File Changes — mechanical replace.
- `backend/go.mod` — adds `pgxmock/v3`.