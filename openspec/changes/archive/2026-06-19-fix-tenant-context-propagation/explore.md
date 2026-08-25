# Exploration: fix-tenant-context-propagation

## Summary

Some backend operations lose the tenant context and produce SQL errors of the form `ERROR: no existe la relación «<table>»` (SQLSTATE 42P01) because the query runs against the `public` schema instead of `tenant_<code>`. The root cause is a **partial / inconsistent fix** in commit `1dc091a` ("feat(backend): add SQL tracer, improve error handling, and fix tenant context propagation") that swapped `c.Context()` → `c.UserContext()` in the **grid handler only**. Eleven other API handlers still call `c.Context()`, which returns the raw `*fasthttp.RequestCtx`. The wrapped Go context that carries the tenant (set via `c.SetUserContext(ctx)` in the tenant middleware) is therefore never seen by `RunInTenantTx` in those code paths.

## Current State — How the tenant is supposed to propagate

### 1. Middleware stores the tenant in TWO places

`backend/internal/infrastructure/middleware/tenant.go` (lines 48–69):

```go
c.Locals(TenantContextKey, tenant)                                              // (a) fiber locals: "tenant" -> "acme"
ctx := context.WithValue(c.Context(), db.TenantContextKey{}, tenant)             // (b) Go context: TenantContextKey{} -> "acme"
c.SetUserContext(ctx)                                                            // (c) fiber user ctx: userContextKey -> wrapped ctx
db.SetTenantSchema(c.Context(), m.pool, tenant)                                  // (d) fasthttp UserValue: "tenant" -> "acme"
```

Four storage locations are populated for every request that carries a tenant. The Fiber handler has multiple ways to retrieve it.

### 2. Fiber's two `Context()` accessors return different things

Verified against `gofiber/fiber/v2@v2.52.0/ctx.go` lines 455–474:

| Method | Returns | Carries tenant? |
|--------|---------|-----------------|
| `c.Context()` | `*fasthttp.RequestCtx` (raw fasthttp request) | The fasthttp `UserValue("tenant")` set by `SetTenantSchema`. Implements `context.Context` (its `Value(key)` → `UserValue(key)`), so `ctx.Value(TenantContextKey{})` hits the **fasthttp UserValue**, not the Go context map. |
| `c.UserContext()` | The `context.Context` previously passed to `c.SetUserContext(ctx)` | The full wrapped `*context.valueCtx` chain with `TenantContextKey{} -> "acme"`. |

### 3. `RunInTenantTx` and `extractTenantCode` (infrastructure/db/context.go)

`RunInTenantTx(ctx, pool, fn)` (line 39) calls `extractTenantCode(ctx)` (line 171) to discover the tenant:

```go
// First try Go context (primary method)
if tenant := ctx.Value(TenantContextKey{}); tenant != nil {  // line 173
    if s, ok := tenant.(string); ok { return s }
}
// Fallback to fasthttp.RequestCtx
if fctx, ok := ctx.(*fasthttp.RequestCtx); ok {              // line 180
    if tenant, ok := fctx.UserValue(tenantKey).(string); ok { return tenant }  // line 181
}
```

The **primary** path (`ctx.Value(TenantContextKey{})`) only succeeds if the ctx that reaches `RunInTenantTx` is the **wrapped** valueCtx (or a child of it). It does **not** succeed on a raw `*fasthttp.RequestCtx`, because fasthttp's `Value(key)` returns `UserValue(key)`, and the fasthttp `UserValue` was never written under `TenantContextKey{}` (the middleware wrote it under Go context's map inside `context.WithValue`, then handed the wrapper to `c.SetUserContext`).

The **fallback** path is intended to save us when the raw `*fasthttp.RequestCtx` reaches `RunInTenantTx` — it looks up `UserValue("tenant")`, which is set by `SetTenantSchema`. The fact that the user's log shows `tenant=''` even on this fallback path is the open question (see Hypothesis H4 below), but it is not blocking: the deterministic fix is to use `c.UserContext()` consistently.

## Affected Areas

### Handlers that use the BROKEN pattern `c.Context()` (must be fixed)

All of these forward a raw `*fasthttp.RequestCtx` to the service layer, bypassing the wrapped tenant context:

| File | Lines with `c.Context()` |
|------|--------------------------|
| `backend/internal/adapters/api/queries/handler.go` | 41, 43, 65, 94, 118, 137 |
| `backend/internal/adapters/api/users/handler.go` | 31, 41, 63, 82, 96 |
| `backend/internal/adapters/api/auth/handler.go` | 32, 59, 79, 99, 118 |
| `backend/internal/adapters/api/stores/handler.go` | 27, 37, 59, 78, 92, 102, 121, 140, 154 |
| `backend/internal/adapters/api/events/handler.go` | 32, 42, 64, 83, 97, 108, 126 |
| `backend/internal/adapters/api/objects/handler.go` | 32, 42, 64, 83, 97, 112 |
| `backend/internal/adapters/api/parts/handler.go` | 32, 42, 64, 83, 97 |
| `backend/internal/adapters/api/stocks/handler.go` | 29, 43, 53, 75, 94, 108, 126, 148, 167, 186, 200 |
| `backend/internal/adapters/api/syscodes/handler.go` | 20, 30, 47, 66, 80, 93 |
| `backend/internal/adapters/api/organizations/handler.go` | 26, 36, 58, 77, 91 |
| `backend/internal/adapters/api/structure/handler.go` | 26, 36, 58, 77, 91 |

Total: **65 call sites** across **11 handler files**.

### Handler that uses the CORRECT pattern `c.UserContext()` (already fixed)

| File | Lines with `c.UserContext()` |
|------|------------------------------|
| `backend/internal/adapters/api/grid/handler.go` | 43, 81, 114 |

This is the only handler that got the fix in commit `1dc091a` ("feat(backend): add SQL tracer, improve error handling, and fix tenant context propagation"). It is the only one for which `POST /api/v1/grid/data` works correctly with a tenant.

### Confirmed working vs. broken paths (from the user report)

| Route | Handler method | Status |
|-------|---------------|--------|
| `POST /api/v1/grid/data` | `GridHandler.ExecuteData` (line 97) | **WORKS** — uses `c.UserContext()` at line 114 |
| `GET /api/v1/queries/:id` | `QueriesHandler.Get` (line 59) | **BROKEN** — uses `c.Context()` at line 65 |
| `GET /api/v1/queries` | `QueriesHandler.List` (line 23) | **BROKEN** — uses `c.Context()` at lines 41, 43 |
| `POST /api/v1/queries` | `QueriesHandler.Create` (line 78) | **BROKEN** — uses `c.Context()` at line 94 |
| `PUT /api/v1/queries/:id` | `QueriesHandler.Update` (line 107) | **BROKEN** — uses `c.Context()` at line 118 |
| `DELETE /api/v1/queries/:id` | `QueriesHandler.Delete` (line 131) | **BROKEN** — uses `c.Context()` at line 137 |

### Layer where the bug is realized

The bug surfaces in `backend/internal/infrastructure/db/context.go` line 51–60 (`RunInTenantTx`), when the empty tenant falls into the `else` branch at line 58–60 and the transaction starts **without** a `SET search_path TO tenant_X, public`. Every SQL statement inside the transaction then runs against `public`, which only has shared system tables — not the per-tenant `eamqueries`, `eamgrids`, `eamstores`, etc.

`backend/internal/infrastructure/db/sql_tracer.go` (line 17) also reads the tenant via the fasthttp fallback. If the fallback is broken, tracer logs lose the tenant too.

## Impact Inventory

### Reads (the user reported these)

- `GET /api/v1/queries/:id` → `eamqueries` SELECT — `ERROR 42P01` (reproduces the user's log)
- `GET /api/v1/queries?gridId=…` and `?gridName=…` → `eamqueries`/`eamgrids` SELECT — same error
- `GET /api/v1/grid/config/:name`, `GET /api/v1/grid/config/id/:id` — these work (uses `c.UserContext()`), but the cross-references in the JOIN go through the same path
- Every other GET handler in the 11 broken files (e.g. `List`, `Get`, `GetLowStock`, `FindAll`, `FindByID`, `FindByOrg`, `FindChildren`, `GetByObject`)

### Writes (HIGHER RISK — silent data corruption into `public`)

- `POST /api/v1/queries` → `INSERT INTO eamqueries (...)` — would insert into the public schema. The next read would fail, but the row is in the wrong schema and may collide with another tenant's row if any tenant has `public.eamqueries` populated.
- `PUT /api/v1/queries/:id` → `UPDATE eamqueries` — would 42P01 (no public table) so still fails, but if a `public.eamqueries` existed (it does not, but the principle stands) it would silently update.
- `DELETE /api/v1/queries/:id` → `DELETE FROM eamqueries` — same.
- All `Create` / `Update` / `Delete` methods in:
  - `users/handler.go` (4 write methods)
  - `stores/handler.go` (8 write methods including bin mutations)
  - `events/handler.go` (4 write methods + status update)
  - `objects/handler.go` (4 write methods)
  - `parts/handler.go` (4 write methods)
  - `stocks/handler.go` (8 write methods including quantity adjust + bin stock mutations)
  - `syscodes/handler.go` (4 write methods)
  - `organizations/handler.go` (4 write methods)
  - `structure/handler.go` (4 write methods)
  - `auth/handler.go` (Register, Logout)

Reads are loud (42P01 surfaces immediately). Writes are louder too because there is no `public.eamqueries` to write to. The actual risk window is narrow because writes also fail loudly — but the failure mode is a 500 error returned to the user, not a clean tenant error.

### Why the auth endpoints also pass `c.Context()` even though they read tenant from `c.Locals`

`auth/handler.go` calls `middleware.GetTenant(c)` (line 20, 47) which reads `c.Locals("tenant")` (defined in `tenant.go:80`). That value is set at the same middleware pass and works correctly. The auth service also receives the tenant string explicitly in the request struct (e.g. `req.Tenant = tenant` at line 30, 57) and then does its own lookup. So auth's tenant is OK for the `public` operations it performs; the `c.Context()` calls in `auth/handler.go` are not dangerous **for the auth flows specifically**, but they are still inconsistent and would break the moment an auth-adjacent code path tries to read/write per-tenant data through `RunInTenantTx`.

## Reproduction

The user has already provided the failing log. Cross-referencing the source:

1. Request: `GET /api/v1/queries/:id`
2. Route: `protected.Get("/queries/:id", c.QueriesHandler.Delete)` — wait, actually: `protected.Delete("/queries/:id", c.QueriesHandler.Delete)` is the DELETE. For GET-by-id, the route is **not currently registered** in `cmd/server/main.go`! Looking at the routes block (lines 149–155), I see:
   - `protected.Get("/queries", c.QueriesHandler.List)` (line 149)
   - `protected.Post("/queries", c.QueriesHandler.Create)` (line 150)
   - `protected.Put("/queries/:id", c.QueriesHandler.Update)` (line 151)
   - `protected.Delete("/queries/:id", c.QueriesHandler.Delete)` (line 152)
   - `protected.Post("/grid/data", c.GridHandler.ExecuteData)` (line 153)
   - `protected.Get("/grid/config/:name", c.GridHandler.GetConfig)` (line 154)
   - `protected.Get("/grid/config/id/:id", c.GridHandler.GetConfigByID)` (line 155)
   
   **There is no `GET /api/v1/queries/:id` route registered.** The `QueriesHandler.Get` method exists (handler.go line 59) but is not wired. So the user's exact failing URL `GET /api/v1/queries/:id` would return 404 from Fiber, not 42P01. The most likely real call paths producing the 42P01 against `eamqueries` are:
   - `GET /api/v1/queries?gridId=…` (List, line 41)
   - `GET /api/v1/queries?gridName=…` (ListByGridName, line 43)
   - Any POST/PUT/DELETE on `/api/v1/queries` (Create, Update, Delete) — but those are writes and would 42P01 on `INSERT/UPDATE/DELETE` not SELECT.
   - Cross-tenant calls: the grid `POST /grid/data` with a `queryId` ends up calling the queries repo (`gridService.ExecuteQueryByID` → `queryRepo.GetByID` → `eamqueries SELECT`); this path works because the grid handler uses `c.UserContext()`.

3. The log lines map directly to:
   - `[TenantMiddleware] Stored tenant='acme' in c.Locals` → `tenant.go:51`
   - `[TenantMiddleware] Set c.SetUserContext with tenant='acme'` → `tenant.go:57`
   - `[SetTenantSchema] tenant='acme' stored in ctx.UserValue` → `context.go:145`
   - `[RunInTenantTx] ctx type=*fasthttp.RequestCtx tenant=''` → `context.go:43` (the `ctx type` part) and `context.go:171` (`extractTenantCode` returning empty)
   - `[TX] [no-tenant] BEGIN (using public schema)` → `context.go:59`
   - `SELECT ... FROM eamqueries WHERE qry_id = $1` → either `query_repository.go:130` (GetByID) or the queries List/ByGridName at lines 28-34 / 82-89
   - `[SQL] <-- ERROR: ERROR: no existe la relación «eamqueries» (SQLSTATE 42P01)` → `pgx` returned this from the SELECT in the no-tenant transaction

## Hypotheses (ranked)

### H1 — HIGH CONFIDENCE: Handlers pass `c.Context()` instead of `c.UserContext()`

This is the most direct explanation and matches both the symptom and the partial history.

- **Evidence**: 11 of 12 API handler files (65 call sites) use `c.Context()`. Only `grid/handler.go` uses `c.UserContext()`. Commit `1dc091a` ("feat(backend): add SQL tracer, improve error handling, and fix tenant context propagation") explicitly states "Fix tenant context propagation: use c.UserContext() instead of c.Context() in handlers" but only changed the grid handler. The grid handler works; the others fail with the exact symptom.
- **Mechanism**: `c.Context()` returns a `*fasthttp.RequestCtx` that does not carry the wrapped Go context. The first lookup in `extractTenantCode` (`ctx.Value(TenantContextKey{})`) returns `nil` because fasthttp's `Value(key)` consults only `UserValue(key)`, and the UserValue at `TenantContextKey{}` was never set. The fallback (`fctx.UserValue("tenant")`) might or might not return the tenant depending on ordering — see H4 — but it is not the deterministic contract the middleware was designed around.
- **Fix shape (proposal territory, not here)**: replace every `c.Context()` in API handlers with `c.UserContext()` and remove the fasthttp fallback in `extractTenantCode` (or keep it as a last-resort safety net). This is mechanical and easy to verify.

### H2 — MEDIUM CONFIDENCE: Even when the fasthttp fallback is in play, ordering is fragile

The fallback in `extractTenantCode` looks at `fctx.UserValue(tenantKey)` where `tenantKey = "tenant"`. This is populated by `db.SetTenantSchema` (line 143) and cleared by `db.ReleaseTenantConn` (line 154). The middleware registers `defer db.ReleaseTenantConn(c.Context())` AFTER calling `SetTenantSchema`, so during the handler the value should be present. But the deferred `ReleaseTenantConn` could be reached on panic/recover paths and remove the value before some other code path sees it. Or a future refactor could re-order the middleware.

- **Evidence**: The user reports `tenant=''` despite the log line `[SetTenantSchema] tenant='acme' stored in ctx.UserValue` having fired. The two should be temporally ordered such that the value is present, but they are not — confirming the fallback is unreliable in practice. Direct proof: the grid handler (which uses `c.UserContext()`) works; the queries handler (which uses `c.Context()`) does not — even though the fallback is the same in both cases.
- **Conclusion**: Even if the fallback works "by accident" for some code paths, it is not the contract and should not be relied on. H1's fix removes the dependency on this fragile ordering entirely.

### H3 — LOW CONFIDENCE: The handler → service → repo chain somewhere replaces the ctx

I checked every service method signature in `internal/domain/*/service.go`. All of them take `ctx context.Context` as the first parameter and pass it through to the repository unchanged. The repository methods all call `infraDB.RunInTenantTx(ctx, r.pool, ...)`. There is no place that calls `context.Background()`, `context.TODO()`, or wraps/replaces the context between the handler and the repository. There is one `go func()` in `cmd/server/main.go:180` (the signal handler that calls `app.Shutdown()`) and one in `cmd/server/main.go:181`-onwards — none of these are tenant-context consumers.

- **Evidence**: grep for `go func(|context.Background|context.TODO` in `backend/` returned only DB connection/migration code (postgres.go, setup/main.go, check/main.go, migrate/main.go) and the shutdown goroutine. None of them are per-request. Searched all domain services for `context.Background/TODO` — none found.
- **Conclusion**: No ctx-swallowing goroutine or wrapper. Eliminated as a primary cause.

### H4 — UNEXPLAINED: Why does the fasthttp fallback return empty when the log says the value was stored?

The middleware logs `[SetTenantSchema] tenant='acme' stored in ctx.UserValue` at `context.go:145` AFTER `ctx.SetUserValue(tenantKey, tenant)` at `context.go:143`. By the time the handler runs, the UserValue should be there. Yet `extractTenantCode` returns empty. Possible explanations (in order of likelihood):

1. **The user pasted logs from different requests mixed together.** The "Stored in c.Locals / Set c.SetUserContext" and "SetTenantSchema" lines might be from one request, and the `RunInTenantTx` log from another (e.g. a request that did not carry a tenant). This is consistent with the user's bug report and easy to test by capturing logs with a request ID.
2. **A bug in fasthttp's `UserValue` lookup under concurrent access.** Unlikely in a single-goroutine handler path, but possible if the fasthttp pool is reset.
3. **The `c.fasthttp` pointer in the handler is different from the one in the middleware.** Would require Fiber to copy the fasthttp ctx, which it does not (verified in `fiber/v2@v2.52.0/ctx.go`).
4. **Something between the middleware and the handler clears the UserValue.** Not found in the codebase.

This is worth flagging to the user during proposal/spec but is not blocking: the deterministic fix is to use `c.UserContext()` regardless. Once H1 is applied, the fasthttp fallback becomes irrelevant and we can either delete it or keep it as defensive code with a clear comment.

## Risks

- **Wide blast radius**: 65 call sites in 11 handler files. The mechanical fix is straightforward, but reviewers must validate that every replacement preserves the surrounding semantics. A pattern like `c.UserContext()` returning a non-nil empty context if never set is well-defined, but the linter/static checks should be tightened.
- **Hidden Go-context contracts**: Some handlers may have other context-derived behavior (deadline, cancellation, request-scoped values from other middlewares). Switching from `c.Context()` to `c.UserContext()` swaps both the value chain and the cancellation/deadline semantics. This is generally desirable (cancellable request-bound ctx) but should be verified handler-by-handler.
- **Tests do not cover this path**: The project config (`openspec/config.yaml`) reports no test files exist. Adding a regression test that asserts `RunInTenantTx` receives a context carrying the tenant for a known-broken route is part of the proposal.
- **The `SetTenantSchema`/`ReleaseTenantConn`/fasthttp fallback code becomes dead code** once all handlers use `c.UserContext()`. We should decide during propose/design whether to remove it (cleaner) or keep it as defensive fallback (safer for future code that may forget).
- **One route was never registered** (`GET /api/v1/queries/:id`). The user's reported URL may be a paraphrase. The proposal should clarify which exact endpoint was failing in the user's reproduction.

## Ready for Proposal

**Yes.** The root cause is unambiguously identified (H1) and the fix surface is bounded (65 call sites, 11 files). Hypotheses H2, H3, H4 are secondary; H1's fix subsumes them. Recommended next phase: `sdd-propose`, with a sharp scope (replace `c.Context()` with `c.UserContext()` in the 11 handler files, plus optional follow-up: remove the fasthttp fallback in `extractTenantCode`, register the missing route, add a regression test).

## Relevant Files

- `backend/internal/infrastructure/middleware/tenant.go` — tenant extraction & ctx wrapping (lines 28–76)
- `backend/internal/infrastructure/db/context.go` — `RunInTenantTx`, `extractTenantCode`, `SetTenantSchema`, `ReleaseTenantConn`
- `backend/internal/infrastructure/db/sql_tracer.go` — also reads tenant via fasthttp fallback
- `backend/internal/adapters/api/grid/handler.go` — the ONLY handler using `c.UserContext()` correctly
- `backend/internal/adapters/api/queries/handler.go` — the failing path in the user's log (uses `c.Context()`)
- `backend/internal/adapters/api/{users,auth,stores,events,objects,parts,stocks,syscodes,organizations,structure}/handler.go` — same bug, 10 more files
- `backend/cmd/server/main.go` — route registration (lines 149–175) and middleware wiring (lines 59–60)
- `gofiber/fiber/v2@v2.52.0/ctx.go` (lines 455–474) — proves `c.Context()` returns raw fasthttp, `c.UserContext()` returns the wrapped one
- `valyala/fasthttp@v1.51.0/server.go` (lines 2715–2755) — proves `*fasthttp.RequestCtx` implements `context.Context` with `Value(key) → UserValue(key)`
- Git history: commit `1dc091a` — the partial fix that touched grid but missed the other 11 handlers
