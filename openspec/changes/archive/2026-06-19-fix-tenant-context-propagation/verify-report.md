## Verification Report

**Change**: `fix-tenant-context-propagation`
**Version**: N/A (no `openspec/specs/` delta — explicit user skip during sdd-spec)
**Mode**: Standard (non-Strict TDD; test is the last of 12 commits)

### Completeness

| Metric | Value |
|--------|-------|
| Tasks total | 12 |
| Tasks complete | 12 |
| Tasks incomplete | 0 |

All 12 tasks in `openspec/changes/fix-tenant-context-propagation/tasks.md` are checked `[x]`. Commits verified via `git log --oneline -12` from `9daad60` (prior tip) to `aadd7f3` (HEAD).

### Build & Tests Execution

**Build**: ✅ Passed
```text
$ cd backend && go build ./...
(exit 0; no output)
```

**Vet**: ✅ Passed
```text
$ cd backend && go vet ./...
(exit 0; no output)
```

**Tests**: ✅ 2/2 passed (full tree has only the new pgxmock test — all other packages report `[no test files]`)
```text
$ cd backend && go test ./...
ok  	github.com/nova/backend/internal/infrastructure/db	(cached)
?   	github.com/nova/backend/cmd/check	[no test files]
?   	github.com/nova/backend/cmd/migrate	[no test files]
... (47 packages, only `internal/infrastructure/db` has tests)
```

**Race detector on db package**: ✅ Passed
```text
$ cd backend && go test -race -v ./internal/infrastructure/db/... -run TestRunInTenantTx_TenantPropagation
=== RUN   TestRunInTenantTx_TenantPropagation
=== RUN   TestRunInTenantTx_TenantPropagation/tenant_wrapped_ctx_emits_set_search_path
2026/06/19 09:44:01 [RunInTenantTx] ctx type=*context.valueCtx tenant='acme'
2026/06/19 09:44:01 [TX] [acme] BEGIN + SET search_path TO tenant_acme, public
2026/06/19 09:44:01 [TX] [acme] COMMIT OK
=== RUN   TestRunInTenantTx_TenantPropagation/no_tenant_uses_public_schema
2026/06/19 09:44:01 [RunInTenantTx] ctx type=context.backgroundCtx tenant=''
2026/06/19 09:44:01 [TX] [no-tenant] BEGIN (using public schema)
2026/06/19 09:44:01 [TX] [no-tenant] COMMIT OK
--- PASS: TestRunInTenantTx_TenantPropagation (0.00s)
    --- PASS: TestRunInTenantTx_TenantPropagation/tenant_wrapped_ctx_emits_set_search_path (0.00s)
    --- PASS: TestRunInTenantTx_TenantPropagation/no_tenant_uses_public_schema (0.00s)
PASS
ok  	github.com/nova/backend/internal/infrastructure/db	1.023s
```

**Coverage**: not enforced by project (`openspec/config.yaml` does not configure coverage). N/A for this change.

### Static Guard (per design §64)

```text
$ grep -rn 'c\.Context()' backend/internal/adapters/api/
backend/internal/adapters/api/grid/handler.go:42:	// Use c.UserContext() instead of c.Context() to get the Go context with tenant
```

Exactly 1 match — a code comment in `grid/handler.go` explicitly excluded by `design.md` line 60: *"The `grid/handler.go:42` comment containing `c.Context()` (a comment, not a call) is NOT in scope."* All 11 originally-broken handlers are clean. ✅

### Manual Smoke (live server)

The prompt asked for `GET /api/v1/queries/1`. That route is **not registered** in `cmd/server/main.go` — this was already flagged in `explore.md` lines 134–143 as out-of-scope. I substituted `GET /api/v1/queries?gridId=1` (the `QueriesHandler.List` method at the originally-broken `handler.go:41`), which IS registered and was the originally failing path that produced the 42P01 error in the user log.

```text
$ curl -s -X POST 'http://localhost:4000/api/v1/auth/login?tenant=acme' \
    -H 'Content-Type: application/json' \
    -d '{"code":"admin","password":"admin123"}'
{"data":{"user":{"id":"1","code":"admin",...},"accessToken":"eyJ...","refreshToken":"rt-...","expiresIn":86400},"success":true}

$ curl -s 'http://localhost:4000/api/v1/queries?gridId=1&tenant=acme' \
    -H "Authorization: Bearer $TOKEN"
{"data":null,"success":true}
HTTP 200
```

**Smoke log evidence — tenant context flowed through `c.UserContext()` correctly into `RunInTenantTx`:**

```text
2026/06/19 09:43:30 [TenantMiddleware] Stored tenant='acme' in c.Locals
2026/06/19 09:43:30 [TenantMiddleware] Set c.SetUserContext with tenant='acme'
2026/06/19 09:43:30 [SQL] --> SET search_path TO tenant_acme, public | args=[]
2026/06/19 09:43:30 [SQL] <-- OK rows=0 cmd=SET
2026/06/19 09:43:30 [SetTenantSchema] tenant='acme' stored in ctx.UserValue
2026/06/19 09:43:30 [RunInTenantTx] ctx type=*context.valueCtx tenant='acme'        ← wrapped Go ctx, NOT *fasthttp.RequestCtx
2026/06/19 09:43:30 [SQL] --> begin | args=[]
2026/06/19 09:43:30 [SQL] <-- OK rows=0 cmd=BEGIN
2026/06/19 09:43:30 [SQL] --> SET search_path TO tenant_acme, public | args=[]     ← per-tenant SET, not public-only
2026/06/19 09:43:30 [SQL] <-- OK rows=0 cmd=SET
2026/06/19 09:43:30 [TX] [acme] BEGIN + SET search_path TO tenant_acme, public     ← per-tenant log line, NOT [no-tenant]
2026/06/19 09:43:30 [SQL] -->  SELECT ... FROM eamqueries WHERE ... | args=[...]
2026/06/19 09:43:30 [SQL] <-- OK rows=0 cmd=SELECT 0
2026/06/19 09:43:30 [SQL] --> commit | args=[]
2026/06/19 09:43:30 [SQL] <-- OK rows=0 cmd=COMMIT
2026/06/19 09:43:30 [TX] [acme] COMMIT OK
09:43:30 | 200 |    7.928404ms | 127.0.0.1 | GET | /api/v1/queries | -
```

Same pattern repeats for the login flow (4–5 additional `[TX] [acme] BEGIN + SET search_path TO tenant_acme, public` blocks when issuing tokens and refreshing sessions). **Every** TX block across the smoke run shows `[acme] …`, never `[no-tenant]`.

The originally-failing `eamqueries` SELECT now resolves cleanly (rows=0 because the grid has no query with `gridId=1`, but no 42P01 schema-not-found error). The bug the user reported is fixed.

### Spec Compliance Matrix

No `openspec/specs/` delta exists for this change (explicit user opt-out during `sdd-propose`, see `proposal.md` "Modified Capabilities" §). Verification therefore relies on the contract spelled out in `design.md §Regression Test Contract` and the smoke-log evidence above.

| Requirement (from design/proposal) | Scenario | Test | Result |
|---|---|---|---|
| `RunInTenantTx` reads tenant from `c.UserContext()`-equivalent wrapped Go ctx and emits `SET search_path TO tenant_<code>, public` | `tenant_wrapped_ctx_emits_set_search_path` | `db.TestRunInTenantTx_TenantPropagation/tenant_wrapped_ctx_emits_set_search_path` | ✅ COMPLIANT (unit) + ✅ confirmed at runtime via smoke (log line `[TX] [acme] BEGIN + SET search_path TO tenant_acme, public`) |
| `RunInTenantTx` does NOT emit `SET search_path` when ctx has no tenant (regression guard) | `no_tenant_uses_public_schema` | `db.TestRunInTenantTx_TenantPropagation/no_tenant_uses_public_schema` | ✅ COMPLIANT (unit) |
| All 11 handlers use `c.UserContext()` instead of `c.Context()` | 70 sites across 11 files | static `grep` + smoke 200 on `/api/v1/queries` (one of the originally-broken routes) | ✅ COMPLIANT |
| Fasthttp fallback in `extractTenantCode` (lines 179–184) remains inert-but-pinned | (DEV-2: explicitly NOT covered by a test) | n/a | ⚠️ PARTIAL — see findings |

**Compliance summary**: 3/4 scenarios compliant. DEV-2 documents the explicit scope reduction.

### Correctness (Static Evidence)

| Requirement | Status | Notes |
|------------|--------|-------|
| 11 handler files use `c.UserContext()` (no stragglers) | ✅ Implemented | grep returns only the explicit in-scope exclusion (grid comment) |
| `RunInTenantTx` signature widened to `QueryEngine` interface | ✅ Implemented | DEV-1: zero runtime impact; `*pgxpool.Pool` already satisfied the interface |
| `context_test.go` covers positive + negative cases | ✅ Implemented | 2 subtests as specified (DEV-2 reduced from design's 3 to match the user prompt) |
| `pgxmock/v3` dependency added | ✅ Implemented | go.mod upgraded to `pgxmock/v3 v3.4.0`; transitive pgx v5.4.3 → v5.5.5 + testify v1.8.1 → v1.9.0 — verified by full-tree build/vet/test |

### Coherence (Design)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| Single PR, one commit per handler file | ✅ Yes | 12 commits on top of `9daad60`; commit-per-file matches tasks.md |
| `pgxmock/v3` (pgx/v5-compatible) | ✅ Yes | `backend/go.mod` imports `github.com/pashagolub/pgxmock/v3` |
| Don't fake `*fiber.Ctx`; pass wrapped ctx directly | ✅ Yes | test uses `context.WithValue(Background(), TenantContextKey{}, "acme")` |
| Keep fasthttp fallback as defensive code | ✅ Yes | `extractTenantCode` lines 183–188 unchanged |
| Test target = `RunInTenantTx` propagation, not fasthttp fallback | ⚠️ Partial | DEV-2 — design wanted 3 subtests incl. fasthttp fallback; apply shipped 2 |
| Test asserts on captured SQL via `mock.ExpectationsWereMet()` | ✅ Yes | `pgxmock.QueryMatcherEqual` matcher, byte-exact comparison |

### Deviations Recap

- **DEV-1 (signature widening: `*pgxpool.Pool` → `QueryEngine`)**: The `QueryEngine` interface in `backend/internal/infrastructure/db/context.go` was **invented in commit 12 (aadd7f3)** — it did NOT exist in commit `1dc091a` (the previous partial fix). At `1dc091a` the file had `func RunInTenantTx(ctx, pool *pgxpool.Pool, ...)`. Commit 12 added the interface (lines 16–24) plus widened the signature to accept it. **→ WARNING per the user's verification prompt.**
- **DEV-2 (test scope reduced from 3 to 2 subtests)**: The fasthttp fallback subtest was dropped. The fallback in `extractTenantCode` (context.go lines 183–188) is now untested; if a future contributor deletes it, no test will fail. **→ WARNING.**
- **DEV-3 (transitive pgx/testify upgrades)**: pgx v5.4.3 → v5.5.5, testify v1.8.1 → v1.9.0. Build/vet/test all pass. **→ SUGGESTION (note in archive).**

### Issues Found

**CRITICAL**: None.

**WARNING**:
- **W1 — `QueryEngine` interface invented in commit 12 (DEV-1)**: Confirmed via `git show 1dc091a -- backend/internal/infrastructure/db/context.go` — the file at the prior tip contained the concrete `*pgxpool.Pool` signature only; `QueryEngine` was added in the same commit that introduced the regression test (`aadd7f3`). This is a design-vs-implementation drift: the change added a new abstraction as part of "the test commit." Mitigation: the interface is minimal (4 methods), trivially satisfied by `*pgxpool.Pool` and `*pgxpool.Conn`, and there are zero callers of `RunInTenantTx` that would break. Future contributors should know the interface is test-only scaffolding so they don't try to "generalize" it into a port.
- **W2 — Fasthttp fallback untested (DEV-2)**: `extractTenantCode` lines 183–188 (`if fctx, ok := ctx.(*fasthttp.RequestCtx); ok { fctx.UserValue(tenantKey) }`) has no regression guard. After this change, the fallback is dead code in practice (no handler passes `c.Context()` anymore), but a future refactor could remove it without any test failing. Mitigation: keep an eye on it; or pin it with a 3rd subtest in a follow-up.
- **W3 — `GET /api/v1/queries/:id` route is not registered** (pre-existing, not caused by this change). The user prompt asked me to hit `GET /api/v1/queries/1`; Fiber returns 404 on that path. The proposed substitute (`GET /api/v1/queries?gridId=1`) is the same handler module (`queries/handler.go:QueriesHandler.List`) at the same originally-failing call site (`c.Context()` at line 41, now `c.UserContext()`). Behavior evidence: HTTP 200 with `SET search_path TO tenant_acme, public` — proves the fix. Registration of `GET /api/v1/queries/:id` is already documented as out-of-scope in `proposal.md §(b)`.

**SUGGESTION**:
- **S1** — Update `proposal.md` "Success Criteria" and `tasks.md` to reflect the actual test contract: 2 subtests, not 3, since DEV-2 explicitly trimmed. Otherwise downstream agents reading the artifacts will look for a missing subtest.
- **S2** — Add `golangci-lint run` to CI for the touched packages. The design lists it as a verification step but the repo has no `.golangci.yml`. Not blocking for archive.
- **S3** — When archiving, the `apply-progress` observation in Engram (id=123) lists DEV-1 as "interface was already declared" — that wording is technically inaccurate (the interface was introduced in the SAME commit as the test, not pre-existing). Archive should correct or simply note that it was effectively introduced.

### Verdict

**PASS WITH WARNINGS**

Build, vet, full test suite, race detector, static guard, and live smoke all succeed. The two WARNINGs (W1 = DEV-1 `QueryEngine` provenance, W2 = DEV-2 untested fasthttp fallback) are documented and accepted; neither breaks the contract. W3 (missing route) is pre-existing and explicitly out of scope. The originally-broken `GET /api/v1/queries?gridId=1` now returns HTTP 200 with the per-tenant `SET search_path TO tenant_acme, public` confirmed in the server log — the user-reported bug is fixed.

### Evidence Files

- Server log snapshot: `/tmp/sdd-verify-logs/server.log` (301 lines, retained for review)
- Regression test file: `backend/internal/infrastructure/db/context_test.go` (88 lines)
- Touched code: `backend/internal/infrastructure/db/context.go` + 11 handler files (full list in `openspec/changes/fix-tenant-context-propagation/tasks.md` Phase 1)