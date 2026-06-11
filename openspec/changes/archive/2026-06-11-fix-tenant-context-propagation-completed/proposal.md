# Proposal: Fix Tenant Context Propagation and Error Handling

## Intent

Fix tenant context propagation so that database queries correctly use the tenant schema (search_path), and improve error handling to provide detailed error responses with SQL error classification.

## Scope

### In Scope
- Fix tenant context propagation from middleware to repository layer
- Ensure all handlers use `c.UserContext()` instead of `c.Context()` for Go context propagation
- Unify context key type to prevent type mismatches between packages
- Improve error handler with SQL error classification (duplicate key, foreign key, syntax errors, etc.)
- Add SQL query tracer for debugging
- Increase JWT expiry from 15 minutes to 24 hours

### Out of Scope
- New features or business logic changes
- Frontend changes unrelated to error handling display
- Database schema changes

## Capabilities

### New Capabilities
- `sql-query-tracer`: pgx QueryTracer logs all SQL queries with tenant context and timing
- `detailed-error-responses`: API returns structured error codes and details for debugging

### Modified Capabilities
- `tenant-context-propagation`: Fixed to work correctly across all handler → service → repository calls
- `jwt-authentication`: Token expiry increased from 15min to 24h

## Approach

### Phase 1: Tenant Context Fix
1. Create unified `TenantContextKey` type in `db/context.go`
2. Update middleware to use `db.TenantContextKey{}` for context.WithValue
3. Update all handlers to use `c.UserContext()` instead of `c.Context()`
4. Verify tenant appears in transaction logs

### Phase 2: Error Handling
1. Enhance `customErrorHandler` in `main.go` to classify errors
2. Add DB error classification (duplicate key, foreign key, syntax, connection)
3. Add detail field to error responses
4. Log internal errors for debugging

### Phase 3: SQL Tracer
1. Create `sql_tracer.go` implementing pgx.QueryTracer
2. Register tracer in `postgres.go` pool config
3. Log all queries with tenant context

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `backend/internal/infrastructure/db/context.go` | Modified | TenantContextKey type, extractTenantCode fix |
| `backend/internal/infrastructure/middleware/tenant.go` | Modified | Use db.TenantContextKey{} |
| `backend/internal/adapters/api/grid/handler.go` | Modified | Use c.UserContext() |
| `backend/cmd/server/main.go` | Modified | Enhanced error handler |
| `backend/internal/infrastructure/db/sql_tracer.go` | Added | SQL query tracer |
| `backend/internal/infrastructure/db/postgres.go` | Modified | Enable tracer |
| `backend/config.yaml` | Modified | JWT expiry 1440 mins |

## Risks

- **Risk**: Changing context propagation could break other handlers not updated
- **Mitigation**: Verify all handlers use c.UserContext() before deploying
- **Risk**: Too much SQL logging could impact performance
- **Mitigation**: Logging is to stdout only; disable in production via log level

## Success Criteria

1. All queries to tenant tables show correct `search_path` in logs
2. `RunInTenantTx` logs show tenant name: `[TX] [acme] BEGIN + SET search_path TO tenant_acme, public`
3. Error responses include `code`, `message`, and `detail` fields
4. SQL queries are logged with tenant context for debugging
5. JWT tokens expire after 24 hours instead of 15 minutes
