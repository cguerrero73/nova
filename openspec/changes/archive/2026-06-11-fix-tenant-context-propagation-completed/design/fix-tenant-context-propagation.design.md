# Design: Fix Tenant Context Propagation and Error Handling

## Overview

This document describes the technical design for fixing tenant context propagation and improving error handling in the Nova backend.

---

## Design1: Unified TenantContextKey

### Problem
Go's `context.WithValue` uses type-based key lookup. If the same key is defined as different types in different packages, `ctx.Value()` returns `nil`.

### Solution
Define `TenantContextKey` as an **exported struct type** in the `db` package. Both `db` and `middleware` packages use the **same type instance**.

### Implementation

**`backend/internal/infrastructure/db/context.go`**:
```go
type TenantContextKey struct{}

func extractTenantCode(ctx context.Context) string {
    if tenant := ctx.Value(TenantContextKey{}); tenant != nil {
        if s, ok := tenant.(string); ok {
            return s
        }
    }
    // Fallback to fasthttp.RequestCtx for backward compat
    if fctx, ok := ctx.(*fasthttp.RequestCtx); ok {
        if tenant, ok := fctx.UserValue(tenantKey).(string); ok {
            return tenant
        }
    }
    return ""
}
```

**`backend/internal/infrastructure/middleware/tenant.go`**:
```go
import "github.com/nova/backend/internal/infrastructure/db"

ctx := context.WithValue(c.Context(), db.TenantContextKey{}, tenant)
c.SetUserContext(ctx)
```

### Why this works
- `db.TenantContextKey{}` is a concrete type
- When middleware stores `ctx.Value(db.TenantContextKey{}, "acme")`, the type is `db.TenantContextKey`
- When `extractTenantCode` reads `ctx.Value(db.TenantContextKey{})`, it uses the **same type**
- Go's context lookup finds the value because the type matches

### Tradeoffs
- Requires handlers to use `c.UserContext()` instead of `c.Context()`
- `c.SetUserContext()` doesn't replace `c.Context()` in Fiber 2.x, so we must use `c.UserContext()` explicitly

---

## Design 2: Handler Context Fix

### Problem
Fiber's `c.Context()` returns `*fasthttp.RequestCtx` which is the original context without our tenant value. `c.SetUserContext()` exists but doesn't replace what `c.Context()` returns.

### Solution
All handlers must use `c.UserContext()` to get the Go context with tenant value.

### Implementation

**Before (broken)**:
```go
func (h *GridHandler) GetConfig(c *fiber.Ctx) error {
    config, err := h.gridService.GetConfig(c.Context(), name)
    // c.Context() returns *fasthttp.RequestCtx - no tenant value
}
```

**After (fixed)**:
```go
func (h *GridHandler) GetConfig(c *fiber.Ctx) error {
    config, err := h.gridService.GetConfig(c.UserContext(), name)
    // c.UserContext() returns Go context with tenant value
}
```

### Files to update
- `backend/internal/adapters/api/grid/handler.go` - all methods
- Any other handler using repositories with tenant context

---

## Design 3: Error Classification

### Problem
`customErrorHandler` returns generic errors without classifying SQL errors or providing actionable details.

### Solution
Enhance error handler to:
1. Check for `*errors.AppError` type
2. Classify database errors by SQLSTATE
3. Return structured `{code, message, detail}` response

### Implementation

```go
func customErrorHandler(c *fiber.Ctx, err error) error {
    code := fiber.StatusInternalServerError
    errCode := "INTERNAL_ERROR"
    message := "An unexpected error occurred"
    var detail string

    // Check for AppError
    if appErr, ok := err.(*errors.AppError); ok {
        code = appErr.GetStatus()
        errCode = appErr.Code
        message = appErr.Message
        detail = appErr.Detail
    }

    // Classify DB errors if detail not already set
    if detail == "" {
        detail = classifyDBError(err)
    }

    // Log internal errors
    if code >= 500 {
        log.Printf("[ERROR] %s | %s | %v", errCode, message, err)
    }

    return c.Status(code).JSON(fiber.Map{
        "success": false,
        "error": fiber.Map{
            "code":    errCode,
            "message": message,
            "detail":  detail,
        },
    })
}

func classifyDBError(err error) string {
    errStr := err.Error()
    switch {
    case containsAny(errStr, "duplicate key", "unique constraint", "23505"):
        return "DUPLICATE_ENTRY"
    case containsAny(errStr, "foreign key", "23503"):
        return "FOREIGN_KEY_VIOLATION"
    case containsAny(errStr, "not null", "23502"):
        return "NOT_NULL_VIOLATION"
    case containsAny(errStr, "connection refused", "dial tcp"):
        return "DATABASE_CONNECTION_ERROR"
    default:
        return ""
    }
}
```

---

## Design 4: SQL Query Tracer

### Problem
No visibility into SQL queries being executed, making debugging difficult.

### Solution
Implement `pgx.QueryTracer` interface to log all queries.

### Implementation

```go
type QueryTracer struct{}

func (t *QueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
    tenant := extractTenantFromCtx(ctx)
    if tenant != "" {
        log.Printf("[SQL] [%s] --> %s | args=%v", tenant, data.SQL, data.Args)
    } else {
        log.Printf("[SQL] --> %s | args=%v", data.SQL, data.Args)
    }
    return ctx
}

func (t *QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
    tenant := extractTenantFromCtx(ctx)
    if data.Err != nil {
        log.Printf("[SQL] <-- ERROR: %v", data.Err)
    } else {
        log.Printf("[SQL] <-- OK rows=%d cmd=%s", data.CommandTag.RowsAffected(), data.CommandTag.String())
    }
}
```

Enable in `postgres.go`:
```go
poolConfig.ConnConfig.Tracer = &QueryTracer{}
```

---

## Design 5: Transaction Logging

### Problem
No visibility into transaction boundaries and search_path settings.

### Solution
Add logging in `RunInTenantTx` for BEGIN, COMMIT, ROLLBACK with tenant context.

### Implementation

```go
func RunInTenantTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
    tenant := extractTenantCode(ctx)

    tx, err := pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("beginning transaction: %w", err)
    }

    if tenant != "" {
        schemaName := fmt.Sprintf("tenant_%s", tenant)
        if _, err := tx.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", schemaName)); err != nil {
            tx.Rollback(ctx)
            return fmt.Errorf("setting search_path: %w", err)
        }
        log.Printf("[TX] [%s] BEGIN + SET search_path TO %s, public", tenant, schemaName)
    } else {
        log.Printf("[TX] [no-tenant] BEGIN (using public schema)")
    }

    if err := fn(tx); err != nil {
        tx.Rollback(ctx)
        log.Printf("[TX] [%s] ROLLBACK (error: %v)", tenant, err)
        return err
    }

    if err := tx.Commit(ctx); err != nil {
        log.Printf("[TX] [%s] COMMIT ERROR: %v", tenant, err)
        return err
    }

    log.Printf("[TX] [%s] COMMIT OK", tenant)
    return nil
}
```

---

## File Summary

| File | Change | Lines |
|------|--------|-------|
| `backend/internal/infrastructure/db/context.go` | TenantContextKey export, extractTenantCode fix | ~30 |
| `backend/internal/infrastructure/middleware/tenant.go` | Use db.TenantContextKey{} | ~10 |
| `backend/internal/adapters/api/grid/handler.go` | c.UserContext() instead of c.Context() | ~10 |
| `backend/cmd/server/main.go` | Enhanced error handler | ~80 |
| `backend/internal/infrastructure/db/sql_tracer.go` | New SQL tracer | ~50 |
| `backend/internal/infrastructure/db/postgres.go` | Enable tracer | ~2 |
| `backend/config.yaml` | JWT expiry 1440 | ~1 |
| **Total** | | **~183** |
