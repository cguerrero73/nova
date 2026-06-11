# Tasks: Fix Tenant Context Propagation and Error Handling

## Overview

Break down the SDD change "fix tenant context propagation and error handling" into discrete, actionable implementation tasks.

---

## Task 1: Export TenantContextKey in db/context.go

**Task ID:** T1  
**Title:** Export TenantContextKey type in db/context.go  
**File(s):** `backend/internal/infrastructure/db/context.go`  
**Status:** [x] Completed

**Description:**
Export `TenantContextKey` as a named struct type in the db package so it can be imported by the middleware package and used consistently.

**Changes:**
1. Replace anonymous `struct{}` or `type contextKey string` with:
```go
type TenantContextKey struct{}
```
2. Update `extractTenantCode` to use `TenantContextKey{}` as the key

**Verification:**
- Type is exported (capital letter)
- Middleware can import and use `db.TenantContextKey{}`

**Estimated Lines:** +5

---

## Task 2: Update Middleware to Use db.TenantContextKey{}

**Task ID:** T2  
**Title:** Update middleware to use db.TenantContextKey{}  
**File(s):** `backend/internal/infrastructure/middleware/tenant.go`  
**Status:** [x] Completed

**Description:**
Update the middleware to import the db package and use `db.TenantContextKey{}` instead of a locally-defined type.

**Changes:**
1. Import `github.com/nova/backend/internal/infrastructure/db`
2. Remove local `tenantContextKey` type definition
3. Use `db.TenantContextKey{}` in `context.WithValue`

**Verification:**
- Code compiles without type mismatch errors
- Tenant value stored with `db.TenantContextKey{}` key

**Estimated Lines:** +5

---

## Task 3: Update Grid Handler to Use c.UserContext()

**Task ID:** T3  
**Title:** Update grid handler to use c.UserContext()  
**File(s):** `backend/internal/adapters/api/grid/handler.go`  
**Status:** [x] Completed

**Description:**
Replace all instances of `c.Context()` with `c.UserContext()` in the grid handler to get the Go context with tenant value.

**Changes:**
1. In `GetConfig`: `h.gridService.GetConfig(c.UserContext(), name)`
2. In `GetConfigByID`: `h.gridService.GetConfigByID(c.UserContext(), gridID)`
3. In `ExecuteData`: `h.gridService.ExecuteQueryByID(c.UserContext(), ...)` and `h.gridService.ExecuteQuery(c.UserContext(), ...)`

**Verification:**
- Handler uses `c.UserContext()` for all repository calls
- Logs show correct tenant in transaction logs

**Estimated Lines:** +6

---

## Task 4: Create SQL Query Tracer

**Task ID:** T4  
**Title:** Create SQL query tracer  
**File(s):** `backend/internal/infrastructure/db/sql_tracer.go` (new)  
**Status:** [x] Completed

**Description:**
Create a pgx QueryTracer implementation that logs all SQL queries with tenant context.

**Changes:**
1. Create new file `sql_tracer.go`
2. Implement `pgx.QueryTracer` interface:
   - `TraceQueryStart` - log SQL and args with tenant
   - `TraceQueryEnd` - log result or error
3. Helper `extractTenantFromCtx` to get tenant from Go context

**Verification:**
- All SQL queries are logged with tenant context
- Errors are logged with full error message

**Estimated Lines:** +50

---

## Task 5: Enable Tracer in PostgresDB

**Task ID:** T5  
**Title:** Enable tracer in postgres.go  
**File(s):** `backend/internal/infrastructure/db/postgres.go`  
**Status:** [x] Completed

**Description:**
Register the SQL tracer in the pgxpool configuration.

**Changes:**
```go
poolConfig.ConnConfig.Tracer = &QueryTracer{}
```

**Verification:**
- Application starts without errors
- SQL queries appear in logs

**Estimated Lines:** +2

---

## Task 6: Enhance Error Handler

**Task ID:** T6  
**Title:** Enhance error handler with classification  
**File(s):** `backend/cmd/server/main.go`  
**Status:** [x] Completed

**Description:**
Enhance `customErrorHandler` to classify SQL errors and return structured error responses.

**Changes:**
1. Check for `*errors.AppError` type
2. Add DB error classification function `classifyDBError`
3. Return `{code, message, detail}` structure
4. Log internal errors for debugging

**Verification:**
- Duplicate key returns `DUPLICATE_ENTRY` code
- Foreign key violation returns `FOREIGN_KEY_VIOLATION` code
- Internal errors are logged

**Estimated Lines:** +80

---

## Task 7: Add Transaction Logging

**Task ID:** T7  
**Title:** Add transaction logging with search_path  
**File(s):** `backend/internal/infrastructure/db/context.go`  
**Status:** [x] Completed

**Description:**
Add logging in `RunInTenantTx` for BEGIN, COMMIT, ROLLBACK with tenant context and search_path.

**Changes:**
1. Log `[TX] [tenant] BEGIN + SET search_path TO tenant_tenant, public`
2. Log `[TX] [tenant] COMMIT OK` on success
3. Log `[TX] [tenant] ROLLBACK (error: ...)` on failure
4. Log `[TX] [no-tenant] BEGIN (using public schema)` when no tenant

**Verification:**
- Transaction boundaries visible in logs
- search_path setting visible in logs

**Estimated Lines:** +20

---

## Task 8: Update JWT Expiry Configuration

**Task ID:** T8  
**Title:** Update JWT expiry to 24 hours  
**File(s):** `backend/config.yaml`  
**Status:** [x] Completed

**Description:**
Change default JWT expiry from 15 minutes to 1440 minutes (24 hours).

**Changes:**
```yaml
jwt:
  expiry_mins: ${JWT_EXPIRY_MINS:-1440}
```

**Verification:**
- Tokens generated during login expire after 24 hours
- Configuration can be overridden via environment variable

**Estimated Lines:** +1

---

## Implementation Order

| Order | Task ID | Title | Risk |
|-------|---------|-------|------|
| 1 | T1 | Export TenantContextKey | Low |
| 2 | T2 | Update middleware | Low |
| 3 | T3 | Update grid handler | Medium |
| 4 | T4 | Create SQL tracer | Low |
| 5 | T5 | Enable tracer | Low |
| 6 | T6 | Enhance error handler | Medium |
| 7 | T7 | Add transaction logging | Low |
| 8 | T8 | Update JWT expiry | Low |

---

## File Summary

| Task | File | Lines | Type |
|------|------|-------|------|
| T1 | `backend/internal/infrastructure/db/context.go` | +5 | Modified |
| T2 | `backend/internal/infrastructure/middleware/tenant.go` | +5 | Modified |
| T3 | `backend/internal/adapters/api/grid/handler.go` | +6 | Modified |
| T4 | `backend/internal/infrastructure/db/sql_tracer.go` | +50 | Added |
| T5 | `backend/internal/infrastructure/db/postgres.go` | +2 | Modified |
| T6 | `backend/cmd/server/main.go` | +80 | Modified |
| T7 | `backend/internal/infrastructure/db/context.go` | +20 | Modified |
| T8 | `backend/config.yaml` | +1 | Modified |
| **Total** | | **~169** | |

---

## Dependencies

- T1 (TenantContextKey export) must be done before T2 (middleware uses it)
- T4 (SQL tracer) should be done before T5 (enable tracer)
- T2 (middleware) enables T3 (handler context fix)

---

## Rollback Summary

| Task | Rollback Action |
|------|-----------------|
| T1 | Revert TenantContextKey to local type definition |
| T2 | Use local tenantContextKey type in middleware |
| T3 | Revert to c.Context() in handlers |
| T4 | Delete sql_tracer.go |
| T5 | Remove Tracer assignment |
| T6 | Revert to simple error handler |
| T7 | Remove transaction logs |
| T8 | Revert JWT expiry to 15 |
