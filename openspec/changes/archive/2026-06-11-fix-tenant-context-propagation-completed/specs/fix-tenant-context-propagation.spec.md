# Spec: Fix Tenant Context Propagation and Error Handling

## Overview

This document specifies the requirements for fixing tenant context propagation and improving error handling in the Nova backend.

---

## Spec 1: Unified TenantContextKey Type

**File**: `backend/internal/infrastructure/db/context.go`

### What
Export a single `TenantContextKey` type used as context key for tenant propagation across all packages.

### Where
- Define `TenantContextKey` struct type in `db/context.go`
- Export type so middleware can import and use it

### Expected Behavior
```go
type TenantContextKey struct{}
```

Both `db` package and `middleware` package use the **same type instance** for context key matching.

### Actual Behavior
- `db/context.go` defines `contextKey string` or anonymous `struct{}`
- `middleware/tenant.go` defines its own `tenantContextKey struct{}`
- Go treats these as different types even with same structure
- `ctx.Value()` returns nil because key type doesn't match

### Fix Description
1. Define `TenantContextKey struct{}` as exported type in `db/context.go`
2. Middleware imports `db` package and uses `db.TenantContextKey{}`
3. `extractTenantCode` uses `TenantContextKey{}` as key

### Test Scenario
```
Scenario: Tenant context propagates correctly
Given a request with X-Tenant-Code: "acme"
When the handler calls c.UserContext()
Then ctx.Value(db.TenantContextKey{}) returns "acme"
And RunInTenantTx logs show "[TX] [acme] BEGIN + SET search_path TO tenant_acme, public"
```

---

## Spec 2: Handler Uses c.UserContext()

**File**: `backend/internal/adapters/api/grid/handler.go`

### What
Change all handlers to use `c.UserContext()` instead of `c.Context()` to get the Go context with tenant value.

### Where
- `GetConfig(c *fiber.Ctx)` at line 42
- `GetConfigByID(c *fiber.Ctx)` at line 77
- `ExecuteData(c *fiber.Ctx)` at lines 112,157

### Expected Behavior
```go
config, err := h.gridService.GetConfig(c.UserContext(), name)
```

### Actual Behavior
```go
config, err := h.gridService.GetConfig(c.Context(), name)
```

`c.Context()` returns `*fasthttp.RequestCtx` which does not contain tenant value set via `context.WithValue`.

### Fix Description
1. Replace all `c.Context()` with `c.UserContext()` in grid handler
2. Apply same change to all other handlers that use repositories

### Test Scenario
```
Scenario: Grid config loads with tenant context
Given tenant "acme" is set in request
When GET /api/v1/grid/config/BMUSER is called
Then the query executes against tenant_acme schema
And logs show "[TX] [acme] BEGIN + SET search_path TO tenant_acme, public"
```

---

## Spec 3: Enhanced Error Handler

**File**: `backend/cmd/server/main.go`

### What
Enhance `customErrorHandler` to classify SQL errors and return structured error responses.

### Where
- `customErrorHandler` function (lines 194-207)

### Expected Behavior
```go
// Error response format:
{
  "success": false,
  "error": {
    "code": "DUPLICATE_ENTRY",
    "message": "User with this email already exists",
    "detail": "duplicate key value violates unique constraint"
  }
}
```

### Actual Behavior
```go
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Error message"
  }
}
```

### Fix Description
1. Check for `*errors.AppError` type and extract code/message/detail
2. Classify database errors by SQLSTATE code:
   - `23505` → `DUPLICATE_ENTRY`
   - `23503` → `FOREIGN_KEY_VIOLATION`
   - `23502` → `NOT_NULL_VIOLATION`
   - `23514` → `CHECK_VIOLATION`
3. Log internal errors (500) for debugging
4. Return empty detail for non-DB errors

### Test Scenario
```
Scenario: Duplicate email returns structured error
Given a user with email "test@example.com" already exists
When POST /api/v1/auth/register is sent with same email
Then response is409 Conflict with:
  code: "DUPLICATE_ENTRY"
  message: "User with this email already exists"
  detail: "duplicate key value violates unique constraint"
```

---

## Spec 4: SQL Query Tracer

**File**: `backend/internal/infrastructure/db/sql_tracer.go` (new)

### What
Create a pgx QueryTracer that logs all SQL queries with tenant context and timing.

### Where
- New file: `backend/internal/infrastructure/db/sql_tracer.go`
- Enable in `backend/internal/infrastructure/db/postgres.go`

### Expected Behavior
```
[SQL] [acme] --> SELECT COUNT(*) FROM eamusers WHERE field = $1 | args=[value]
[SQL] [acme] <-- OK rows=100 cmd=SELECT
[SQL] [acme] <-- ERROR: relation "bad_table" does not exist
```

### Actual Behavior
No SQL query logging exists.

### Fix Description
1. Create `QueryTracer` struct implementing `pgx.QueryTracer` interface
2. Implement `TraceQueryStart` to log SQL and args with tenant
3. Implement `TraceQueryEnd` to log result or error
4. Register tracer in `pgxpool.ParseConfig.ConnConfig.Tracer`

### Test Scenario
```
Scenario: SQL queries are logged with tenant
Given tenant "acme" is set
When any database query is executed
Then logs show "[SQL] [acme] --> SQL | args=[...]"
And logs show "[SQL] [acme] <-- OK rows=N cmd=..."
```

---

## Spec 5: JWT Expiry Configuration

**File**: `backend/config.yaml`

### What
Change default JWT expiry from 15 minutes to 1440 minutes (24 hours).

### Where
- Line 11: `expiry_mins: ${JWT_EXPIRY_MINS:-15}`

### Expected Behavior
```yaml
jwt:
  expiry_mins: ${JWT_EXPIRY_MINS:-1440}
```

### Actual Behavior
```yaml
jwt:
  expiry_mins: ${JWT_EXPIRY_MINS:-15}
```

### Fix Description
1. Change default value from 15 to 1440
2. No code changes needed; config value is already used correctly

### Test Scenario
```
Scenario: JWT token has24h expiry
Given JWT is generated during login
When the token is decoded
Then exp claim is 24 hours from iat
```
