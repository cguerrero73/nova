# Tasks: Fix Tenant Context Propagation Across API Handlers

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~250–350 (70 mechanical replacements across 11 files + ~80–150 test file) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR, 12 reviewable work-unit commits |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending (single PR — not applicable) |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Work-unit commits (12 commits, 1 PR)

| # | Commit title | Files touched | Net diff |
|---|--------------|---------------|----------|
| 1 | `fix(tenant): use c.UserContext() in queries handler` | `adapters/api/queries/handler.go` | ~6 sites |
| 2 | `fix(tenant): use c.UserContext() in users handler` | `adapters/api/users/handler.go` | ~5 sites |
| 3 | `fix(tenant): use c.UserContext() in auth handler` | `adapters/api/auth/handler.go` | ~5 sites |
| 4 | `fix(tenant): use c.UserContext() in stores handler` | `adapters/api/stores/handler.go` | ~9 sites |
| 5 | `fix(tenant): use c.UserContext() in events handler` | `adapters/api/events/handler.go` | ~7 sites |
| 6 | `fix(tenant): use c.UserContext() in objects handler` | `adapters/api/objects/handler.go` | ~6 sites |
| 7 | `fix(tenant): use c.UserContext() in parts handler` | `adapters/api/parts/handler.go` | ~5 sites |
| 8 | `fix(tenant): use c.UserContext() in stocks handler` | `adapters/api/stocks/handler.go` | ~11 sites |
| 9 | `fix(tenant): use c.UserContext() in syscodes handler` | `adapters/api/syscodes/handler.go` | ~6 sites |
| 10 | `fix(tenant): use c.UserContext() in organizations handler` | `adapters/api/organizations/handler.go` | ~5 sites |
| 11 | `fix(tenant): use c.UserContext() in structure handler` | `adapters/api/structure/handler.go` | ~5 sites |
| 12 | `test(tenant): regression test for RunInTenantTx context propagation` | `infrastructure/db/context_test.go` (new), `go.mod`, `go.sum` | ~80–150 lines |

Rollback (applies globally to every task): each commit is revertible with `git revert <sha>`. Because tasks 1–11 touch one file each, a per-task revert removes that handler's fix without affecting siblings. Task 12 is additive and reverts cleanly.

---

## Phase 1: Mechanical Handler Fixes (Tasks 1–11)

Each task is a single commit, one handler file, identical shape: every `c.Context()` becomes `c.UserContext()`. Per-task counts are verified against the actual files at `openspec/changes/fix-tenant-context-propagation/design.md` time and re-verified by `sdd-tasks` (see Risks if any count drifts).

Common verification for tasks 1–11:

```bash
cd backend
go build ./...
# Static guard — must be 0 matches in the touched file
grep -c 'c\.Context()' internal/adapters/api/<scope>/handler.go
# Spot-check that the wrapped ctx now reaches RunInTenantTx
grep -n 'UserContext' internal/adapters/api/<scope>/handler.go
```

Common do-not-touch notes for tasks 1–11:

- `context.Background()`, `context.WithValue(...)`, `context.TODO()` — package-qualified, NOT Fiber. Must stay.
- No string-literal `"context."` exists in the 11 files (audited).
- Zero `*fasthttp.RequestCtx` type assertions across the 11 files (audited). The swap is type-safe; `c.UserContext()` returns `context.Context`, same as `c.Context()`.
- The `grid/handler.go:42` comment containing `c.Context()` (a comment, not a call) is NOT in scope.

### Task 1 — [x] `fix(tenant): use c.UserContext() in queries handler`

- **File**: `backend/internal/adapters/api/queries/handler.go` (6 sites, lines 41, 43, 65, 94, 118, 137)
- **Diff sketch**:
  ```diff
  - result, err = h.queryService.List(c.Context(), gridID, userCode)
  + result, err = h.queryService.List(c.UserContext(), gridID, userCode)
  ```
- **Verification**: `go build ./...`; `grep -c 'c\.Context()' backend/internal/adapters/api/queries/handler.go` returns `0`.

### Task 2 — [x] `fix(tenant): use c.UserContext() in users handler`

- **File**: `backend/internal/adapters/api/users/handler.go` (5 sites, lines 31, 41, 63, 82, 96)
- **Diff sketch**: identical pattern as Task 1 (global `c.Context()` → `c.UserContext()`).
- **Verification**: `go build ./...`; per-file count returns `0`.

### Task 3 — [x] `fix(tenant): use c.UserContext() in auth handler`

- **File**: `backend/internal/adapters/api/auth/handler.go` (5 sites, lines 32, 59, 79, 99, 118)
- **Diff sketch**: identical pattern.
- **Note**: Auth currently reads tenant via `middleware.GetTenant(c)` (c.Locals), so the swap is not load-bearing for current auth flows — it is consistent with the rest of the codebase and protects future auth-adjacent code paths.
- **Verification**: `go build ./...`; per-file count returns `0`.

### Task 4 — [x] `fix(tenant): use c.UserContext() in stores handler`

- **File**: `backend/internal/adapters/api/stores/handler.go` (9 sites, lines 27, 37, 59, 78, 92, 102, 121, 140, 154)
- **Diff sketch**: identical pattern.
- **Verification**: `go build ./...`; per-file count returns `0`.

### Task 5 — [x] `fix(tenant): use c.UserContext() in events handler`

- **File**: `backend/internal/adapters/api/events/handler.go` (7 sites, lines 32, 42, 64, 83, 97, 108, 126)
- **Diff sketch**: identical pattern.
- **Verification**: `go build ./...`; per-file count returns `0`.

### Task 6 — [x] `fix(tenant): use c.UserContext() in objects handler`

- **File**: `backend/internal/adapters/api/objects/handler.go` (6 sites, lines 32, 42, 64, 83, 97, 112)
- **Diff sketch**: identical pattern.
- **Verification**: `go build ./...`; per-file count returns `0`.

### Task 7 — [x] `fix(tenant): use c.UserContext() in parts handler`

- **File**: `backend/internal/adapters/api/parts/handler.go` (5 sites, lines 32, 42, 64, 83, 97)
- **Diff sketch**: identical pattern.
- **Verification**: `go build ./...`; per-file count returns `0`.

### Task 8 — [x] `fix(tenant): use c.UserContext() in stocks handler`

- **File**: `backend/internal/adapters/api/stocks/handler.go` (11 sites, lines 29, 43, 53, 75, 94, 108, 126, 148, 167, 186, 200)
- **Diff sketch**: identical pattern.
- **Note**: Largest single commit. Still ~11 line-level changes; diff is small even at the file level.
- **Verification**: `go build ./...`; per-file count returns `0`.

### Task 9 — [x] `fix(tenant): use c.UserContext() in syscodes handler`

- **File**: `backend/internal/adapters/api/syscodes/handler.go` (6 sites, lines 20, 30, 47, 66, 80, 93)
- **Diff sketch**: identical pattern.
- **Verification**: `go build ./...`; per-file count returns `0`.

### Task 10 — [x] `fix(tenant): use c.UserContext() in organizations handler`

- **File**: `backend/internal/adapters/api/organizations/handler.go` (5 sites, lines 26, 36, 58, 77, 91)
- **Diff sketch**: identical pattern.
- **Verification**: `go build ./...`; per-file count returns `0`.

### Task 11 — [x] `fix(tenant): use c.UserContext() in structure handler`

- **File**: `backend/internal/adapters/api/structure/handler.go` (5 sites, lines 26, 36, 58, 77, 91)
- **Diff sketch**: identical pattern.
- **Verification**: `go build ./...`; per-file count returns `0`.

After all 11 commits:

```bash
# Global guard: 0 stragglers across all API handlers
grep -rn 'c\.Context()' backend/internal/adapters/api/ | grep -v 'c\.UserContext()'
# Expect: no output
```

---

## Phase 2: Regression Test (Task 12)

### Task 12 — [x] `test(tenant): regression test for RunInTenantTx context propagation`

- **Files**:
  - **New**: `backend/internal/infrastructure/db/context_test.go`
  - **Modified**: `backend/go.mod`, `backend/go.sum` (add `github.com/pashagolub/pgxmock/v3`)
- **Pre-step** (must run before writing the test file):
  ```bash
  cd backend
  go get github.com/pashagolub/pgxmock/v3
  go mod tidy
  ```
  Use the **`/v3`** import path. The unversioned path (`pashagolub/pgxmock`) is v2/pgx/v4 and will not compile against `pgx/v5`.

- **Test package**: `package db` (white-box; same package as `context.go` so `extractTenantCode` is reachable).

- **Imports** (suggested):
  ```go
  import (
      "context"
      "testing"

      "github.com/jackc/pgx/v5"
      "github.com/pashagolub/pgxmock/v3"
      "github.com/stretchr/testify/require"
      "github.com/valyala/fasthttp"
  )
  ```
  `stretchr/testify` is already a Go-ecosystem standard; if the repo has not adopted it, fall back to plain `t.Fatal/t.Errorf` (no new dep needed). `pgx/v5` is already in `go.mod`.

- **Test function name**: `TestRunInTenantTx_TenantPropagation`

- **Subtests** (table-driven, three cases; each asserts on captured SQL via `mock.ExpectationsWereMet()`):

  | Subtest | `ctx` value | Expected SQL (in order) | Why |
  |---------|-------------|-------------------------|-----|
  | `tenant_wrapped_ctx` | `context.WithValue(context.Background(), TenantContextKey{}, "acme")` | `SET search_path TO tenant_acme, public` then a `SELECT 1` inside the user fn | Positive contract — `c.UserContext()`-equivalent wrapped ctx carries the tenant |
  | `tenant_in_fasthttp_user_value` | `&fasthttp.RequestCtx{}` with `SetUserValue("tenant", "legacy")` | `SET search_path TO tenant_legacy, public` then `SELECT 1` | Pins the legacy fallback so any future removal is intentional |
  | `no_tenant` | `context.Background()` | `SELECT 1` only (no `SET search_path`) | Pins the no-tenant / public-schema branch |

  Per subtest skeleton (NOT the final body — sdd-apply writes it):
  ```go
  for _, tt := range tests {
      t.Run(tt.name, func(t *testing.T) {
          mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
          require.NoError(t, err)
          defer mock.Close()

          mock.ExpectBegin()
          if tt.expectSetSchema != "" {
              mock.ExpectExec(tt.expectSetSchema).WillReturnResult(pgconn.NewCommandTag("SET"))
          }
          mock.ExpectExec("SELECT 1").WillReturnResult(pgconn.NewCommandTag("SELECT 1"))
          mock.ExpectCommit()

          err = RunInTenantTx(tt.ctx, mock, func(tx pgx.Tx) error {
              _, err := tx.Exec(tt.ctx, "SELECT 1")
              return err
          })
          require.NoError(t, err)
          require.NoError(t, mock.ExpectationsWereMet())
      })
  }
  ```

- **What the test must NOT do**:
  - Must NOT instantiate a real `*fiber.Ctx` — `RunInTenantTx` only reads its `ctx` parameter.
  - Must NOT call `SetTenantSchema` / `ReleaseTenantConn` (those are fasthttp lifecycle helpers, not relevant to `RunInTenantTx`).
  - Must NOT register or hit any HTTP route — pure unit test, no server boot.

- **Verification**:
  ```bash
  cd backend
  go build ./...
  go test ./internal/infrastructure/db/... -v -run TestRunInTenantTx_TenantPropagation
  go vet ./internal/infrastructure/db/...
  ```

- **Notes for sdd-apply**:
  - This is the **first `*_test.go`** in the repo. `go mod tidy` and `go test` must both succeed before declaring this task done.
  - If `pgxmock.QueryMatcherEqual` causes friction with how `RunInTenantTx` formats the `SET search_path` (via `fmt.Sprintf`), verify the exact string the mock expects vs. what `RunInTenantTx` emits — match by string equality.
  - `extractTenantCode` is package-private; `package db` placement is required.

---

## Phase 3: Post-Apply Global Verification (sdd-verify)

The sdd-verify phase will run:

```bash
cd backend

# 1. Static guard — 0 stragglers across all API handlers
grep -rn 'c\.Context()' internal/adapters/api/ | grep -v 'c\.UserContext()'
# Expect: no output

# 2. Compile
go build ./...

# 3. Vet (catches type-mismatch regressions at the ctx accessor swap)
go vet ./...

# 4. Regression test
go test ./internal/infrastructure/db/... -v -run TestRunInTenantTx_TenantPropagation

# 5. Lint (if configured)
golangci-lint run ./internal/adapters/api/... ./internal/infrastructure/db/...
```

Manual smoke (out of scope for CI, documented in `design.md` §Verification):

```bash
curl -H "X-Tenant-Code: acme" 'http://localhost:8080/api/v1/queries?gridId=1'
# Expect: 200 + rows; log shows `[TX] [acme] BEGIN + SET search_path TO tenant_acme, public`
```

---

## Implementation Order

Tasks 1–11 are independent and could run in any order. Recommended order follows the user's listing (queries first because it is the reported failing path). Task 12 goes **last** so the test pins the contract after the production code is already correct — if the test is written first, it passes only because the handlers are still wrong (no `UserContext` to assert on). Running the test last guarantees it captures the *fixed* behavior.

Each task is sized for one focused commit and is independently revertible.

---

## Out of Scope (explicit)

- Registering the missing `GET /api/v1/queries/:id` route (proposal §Out-of-scope (b)).
- Removing the fasthttp fallback in `extractTenantCode` (kept as defensive code; third subtest in task 12 locks its behavior).
- Any refactor of the tenant middleware, service layer, or repositories.
- Delta spec / `openspec/specs/` — `openspec/specs/` does not exist in this repo; user explicitly skipped sdd-spec. The fix is mechanical and the regression test pins behavior.

---

## Relevant Files

- `backend/internal/infrastructure/db/context.go` — `RunInTenantTx` (line 39), `extractTenantCode` (line 171). Test target.
- `backend/internal/infrastructure/middleware/tenant.go` — wraps ctx at lines 55–57. NOT modified.
- `backend/internal/adapters/api/grid/handler.go` — canonical `c.UserContext()` reference. NOT modified.
- 11 handlers in tasks 1–11 — mechanical replace.
- `backend/go.mod` / `backend/go.sum` — add `github.com/pashagolub/pgxmock/v3`.
- `backend/internal/infrastructure/db/context_test.go` — new file in task 12.