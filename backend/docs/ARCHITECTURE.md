# Backend Architecture Documentation

## Overview

Nova backend is a Go application using Fiber framework with pgx for PostgreSQL connectivity. It implements multi-tenant isolation using PostgreSQL schemas and transaction-scoped `search_path`.

---

## Table of Contents

1. [Architecture](#architecture)
2. [Startup](#startup)
3. [Request Lifecycle](#request-lifecycle)
4. [Multi-Tenant Isolation](#multi-tenant-isolation)
5. [Transaction Mode](#transaction-mode)
6. [PgBouncer Integration](#pgbouncer-integration)
7. [Directory Structure](#directory-structure)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         GO APPLICATION                          │
│                                                                 │
│  cmd/server/main.go                                            │
│       │                                                        │
│       ├── Config (env vars)                                    │
│       ├── Wire (dependency injection)                          │
│       │      └── Creates pgxpool                               │
│       │                                                        │
│       └── Fiber App                                            │
│            ├── Global Middleware                                │
│            │   └── ExtractTenant                                │
│            │                                                    │
│            └── Routes → Handlers → Services → Repositories     │
│                              │                   │               │
│                              │                   └── db/context   │
│                              │                       └── RunInTenantTx
│                              │                                            │
│                              └── middleware/auth.go               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    PostgreSQL                                    │
│                                                                 │
│  public (global schema)                                         │
│  ├── eamtenants                                                 │
│  └── ...                                                        │
│                                                                 │
│  tenant_acme                                                    │
│  ├── eamusers                                                   │
│  ├── eamobjects                                                 │
│  ├── eamorganizations                                           │
│  └── ...                                                        │
│                                                                 │
│  tenant_xyz                                                     │
│  └── ...                                                        │
└─────────────────────────────────────────────────────────────────┘
```

---

## Startup

### Flow

```
1. main.go
   ├── Load configuration from env vars
   │   ├── DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_DATABASE
   │   ├── JWT_SECRET, JWT_EXPIRY_MINS
   │   └── SERVER_PORT
   │
   ├── Initialize pgxpool
   │   └── pgxpool.New(context.Background(), DATABASE_URL)
   │
   ├── Wire dependencies
   │   ├── auth.Service (UserRepo, SessionRepo, JWTConfig)
   │   ├── users.Service (UserRepo)
   │   ├── objects.Service (ObjectRepo)
   │   └── ... (one service per domain)
   │
   ├── Create Fiber app
   │   └── app := fiber.New()
   │
   ├── Register global middleware
   │   └── app.Use(TenantMiddleware.ExtractTenant())
   │
   ├── Register routes
   │   ├── /api/v1/auth/* → AuthHandler
   │   ├── /api/v1/users/* → UserHandler
   │   ├── /api/v1/objects/* → ObjectHandler
   │   └── ... (one handler per domain)
   │
   └── Start server
       └── app.Listen(":4000")
```

### Configuration (config.yaml)

```yaml
database:
  host: ${DB_HOST:-localhost}
  port: ${DB_PORT:-5432}
  user: ${DB_USER:-postgres}
  password: ${DB_PASSWORD:-postgres}
  database: ${DB_DATABASE:-nova}
  schema: ${DB_SCHEMA:-public}

jwt:
  secret: ${JWT_SECRET:-your-super-secret-key-change-in-production}
  expiry_mins: ${JWT_EXPIRY_MINS:-15}

server:
  port: ${SERVER_PORT:-4000}
  env: ${SERVER_ENV:-development}
```

### Running

```bash
# Development
make dev

# With custom config
DB_HOST=192.168.1.100 DB_PORT=5432 make dev

# Production (via Makefile)
make build
./nova-server
```

---

## Request Lifecycle

### Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  REQUEST: POST /api/v1/auth/login?tenant=acme                               │
│  Body: { "code": "admin", "password": "..." }                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  MIDDLEWARE: ExtractTenant                                                   │
│                                                                             │
│  1. Extract tenant from:                                                    │
│     - Query param:  ?tenant=acme                                             │
│     - Header:     X-Tenant-Code: acme                                       │
│     - Body:        { "tenant": "acme", ... }                                │
│                                                                             │
│  2. Store in fasthttp context:                                              │
│     c.Context().SetUserValue("tenant", "acme")                              │
│                                                                             │
│  3. Pass to next handler:                                                   │
│     return c.Next()                                                         │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  HANDLER: AuthHandler.Login()                                               │
│                                                                             │
│  1. Get tenant from context:                                                │
│     tenant := middleware.GetTenant(c)  // → "acme"                          │
│                                                                             │
│  2. Parse request:                                                          │
│     var req auth.LoginRequest                                                │
│     c.BodyParser(&req)                                                      │
│     req.Tenant = tenant                                                     │
│                                                                             │
│  3. Call service:                                                           │
│     resp, err := h.authService.Login(c.Context(), &req)                    │
│                                                                             │
│  4. Return response:                                                         │
│     return c.JSON(fiber.Map{"success": true, "data": resp})                 │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  SERVICE: AuthService.Login()                                                │
│                                                                             │
│  1. Find user by code:                                                      │
│     user, err := s.userRepo.FindByCode(ctx, req.Code)                      │
│                                                                             │
│  2. Validate password:                                                      │
│     bcrypt.CompareHashAndPassword(user.Password, req.Password)              │
│                                                                             │
│  3. Generate auth response:                                                  │
│     return s.generateAuthResponse(ctx, user)                                │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  REPOSITORY: PgUserRepository.FindByCode()                                   │
│                                                                             │
│  err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {          │
│                                                                             │
│      // Query executes here (with search_path set)                          │
│      query := `SELECT * FROM eamusers WHERE usr_code = $1`                 │
│      return tx.QueryRow(ctx, query, code).Scan(&user)                       │
│                                                                             │
│  })                                                                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  HELPERS: RunInTenantTx()                                                    │
│                                                                             │
│  func RunInTenantTx(ctx context.Context, pool *pgxpool.Pool,                 │
│                     fn func(tx pgx.Tx) error) error {                       │
│                                                                             │
│      // 1. Extract tenant from context                                      │
│      tenant := extractTenantCode(ctx)  // → "acme"                          │
│                                                                             │
│      // 2. Begin transaction                                                 │
│      tx, err := pool.Begin(ctx)                                             │
│      if err != nil { return err }                                           │
│                                                                             │
│      // 3. Set search_path (only if tenant exists)                          │
│      if tenant != "" {                                                      │
│          schemaName := "tenant_" + tenant  // → "tenant_acme"                │
│          tx.Exec(ctx, "SET search_path TO "+schemaName+", public")           │
│      }                                                                      │
│                                                                             │
│      // 4. Execute repository queries                                       │
│      if err := fn(tx); err != nil {                                         │
│          tx.Rollback(ctx)                                                    │
│          return err                                                         │
│      }                                                                      │
│                                                                             │
│      // 5. Commit transaction                                               │
│      return tx.Commit(ctx)                                                 │
│  }                                                                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  RESPONSE: { "success": true, "data": { "token": "...", "user": {...} } }   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Multi-Tenant Isolation

### How It Works

Each tenant has its own PostgreSQL schema:

- Tenant `acme` → schema `tenant_acme`
- Tenant `xyz` → schema `tenant_xyz`

The `search_path` is set per-transaction, not per-connection.

### Schema Structure

```
PostgreSQL
│
├── public (global)
│   └── eamtenants (tenant registry)
│
├── tenant_acme (isolated)
│   ├── eamusers
│   ├── eamobjects
│   ├── eamorganizations
│   ├── eamparts
│   ├── eamstores
│   ├── eambins
│   ├── eamstocks
│   ├── eamevents
│   └── eamstructure
│
└── tenant_xyz (isolated)
    └── ... (same tables)
```

### Isolation Guarantees

| Guarantee                    | Implementation                                |
| ---------------------------- | --------------------------------------------- |
| Tenant data isolated         | PostgreSQL schema per tenant                  |
| Queries go to correct schema | `SET search_path TO tenant_X` per transaction |
| No cross-tenant leaks        | Schema-level isolation                        |
| Tenant created correctly     | Migration creates both entry + schema         |

### Creating a New Tenant

```bash
# Via Makefile
make create-tenant TENANT=acme

# SQL executed:
INSERT INTO eamtenants (ten_code, ten_name) VALUES ('acme', 'acme');
CREATE SCHEMA tenant_acme;
```

---

## Transaction Mode

### Why Transactions?

In session mode, each request holds a database connection for its entire duration. With many concurrent users, this can exhaust PostgreSQL's connection limit.

Transaction mode allows PgBouncer to multiplex many logical connections over fewer physical connections.

### How It Works

```
BEFORE (Session Mode):
┌──────────┐     ┌──────────┐     ┌──────────┐
│  User 1  │     │  User 2  │     │  User N  │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     ▼                ▼                ▼
┌─────────────────────────────────────────┐
│         PostgreSQL Connections           │
│  conn1 (holds search_path=tenant_acme)   │
│  conn2 (holds search_path=tenant_xyz)     │
│  conn3 ...                                │
│  ...                                      │
│  If N=100 users, need 100 connections     │
└─────────────────────────────────────────┘

AFTER (Transaction Mode):
┌──────────┐     ┌──────────┐     ┌──────────┐
│  User 1  │     │  User 2  │     │  User N  │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     ▼                ▼                ▼
┌─────────────────────────────────────────┐
│     PgBouncer (transaction mode)        │
│  ┌────┐ ┌────┐ ┌────┐                   │
│  │pool│ │pool│ │pool│ (20 connections)  │
│  └──┬─┘ └──┬─┘ └──┬─┘                   │
└─────┼──────┼──────┼─────────────────────┘
      │       │       │
      ▼       ▼       ▼
┌─────────────────────────────────────────┐
│     PostgreSQL (20 physical connections)  │
└─────────────────────────────────────────┘
```

### Transaction Lifecycle with search_path

```sql
BEGIN;
  SET search_path TO tenant_acme, public;
  SELECT * FROM eamusers WHERE ...;  -- searches in tenant_acme
  INSERT INTO eamobjects ...;         -- inserts into tenant_acme
  SELECT * FROM eamorganizations...;  -- searches in tenant_acme
COMMIT;
-- Transaction ends, search_path is lost but that's fine
```

### Key Properties

| Property                   | Description                                      |
| -------------------------- | ------------------------------------------------ |
| One transaction            | One logical operation (login, save object, etc.) |
| search_path set once       | At the beginning of each transaction             |
| All queries in same schema | The entire transaction uses the same tenant      |
| Commit or rollback         | Atomic operation                                 |

### Example: Multiple Queries in One Transaction

```go
func (s *ObjectService) CreateWithHistory(ctx context.Context, obj *Object) error {
    return infraDB.RunInTenantTx(ctx, s.pool, func(tx pgx.Tx) error {
        // 1. Create object
        if err := tx.Exec(ctx, "INSERT INTO eamobjects ...", obj...); err != nil {
            return err  // automatic rollback
        }

        // 2. Create history entry
        if err := tx.Exec(ctx, "INSERT INTO eamhistory ...", ...); err != nil {
            return err  // automatic rollback
        }

        // 3. Update counter (all in tenant_acme)
        if err := tx.Exec(ctx, "UPDATE counters SET ...", ...); err != nil {
            return err  // automatic rollback
        }

        return nil  // commit
    })
}
```

---

## PgBouncer Integration

### Architecture with PgBouncer

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        PRODUCTION SETUP                                 │
└─────────────────────────────────────────────────────────────────────────┘

┌────────────────┐                           ┌────────────────────────────┐
│   GO SERVER     │                           │     DB SERVER              │
│   (your app)    │                           │                            │
│                 │                           │  ┌────────────────────┐  │
│  DB_HOST=ip     │      TCP/IP              │  │    PgBouncer        │  │
│  DB_PORT=6432   │ ◄────────────────────► │  │    port: 6432       │  │
│                 │                           │  │    pool_mode=session│  │
│  pgxpool ──────┼───────────────────────────┼  │    default_pool=20  │  │
│  (connection    │                           │  └─────────┬──────────┘  │
│   pool in app)  │                           │            │             │
│                 │                           │  ┌─────────▼──────────┐  │
│                 │                           │  │    PostgreSQL      │  │
│                 │                           │  │    port: 5432      │  │
│                 │                           │  └────────────────────┘  │
└────────────────┘                           └────────────────────────────┘
```

### Configuration

**pgbouncer.ini** (on DB server):

```ini
[databases]
nova = host=127.0.0.1 port=5432 dbname=nova user=dev password=dev

[pgbouncer]
listen_port = 6432
listen_addr = 0.0.0.0

pool_mode = session
max_client_conn = 100
default_pool_size = 20

query_timeout = 60
query_wait_timeout = 30
server_lifetime = 3600
server_idle_timeout = 600
```

### Key Settings

| Setting             | Value     | Purpose                              |
| ------------------- | --------- | ------------------------------------ |
| `pool_mode`         | `session` | Required for search_path persistence |
| `default_pool_size` | `20`      | Physical connections to PostgreSQL   |
| `max_client_conn`   | `100`     | Maximum logical connections from Go  |
| `server_lifetime`   | `3600`    | Recycle connections after 1 hour     |

### Why Session Mode?

Transaction mode would NOT work because:

```
Transaction mode:
  BEGIN
    SET search_path TO tenant_acme
  COMMIT  ← PgBouncer returns connection to pool, search_path lost!

  BEGIN (different user, same conn)
    SELECT * FROM users  ← goes to wrong schema!

Session mode:
  BEGIN
    SET search_path TO tenant_acme
    SELECT * FROM users
    COMMIT

  NEXT REQUEST:
    BEGIN (same conn, same user)
      SET search_path TO tenant_acme  ← fresh start
      SELECT * FROM users
    COMMIT
```

### Benefits

| Metric                 | Without PgBouncer | With PgBouncer |
| ---------------------- | ----------------- | -------------- |
| PostgreSQL connections | 1 per user        | 20 fixed       |
| Scale limit            | ~100 users        | 100+ users     |
| Connection overhead    | High              | Low            |

---

## Directory Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
│
├── internal/
│   ├── adapters/
│   │   ├── api/                 # HTTP handlers
│   │   │   ├── auth/            # Auth handler
│   │   │   ├── users/           # Users handler
│   │   │   ├── objects/         # Objects handler
│   │   │   └── .../            # Other handlers
│   │   │
│   │   └── db/                  # Database repositories
│   │       ├── auth/
│   │       │   ├── user_repository.go
│   │       │   └── session_repository.go
│   │       ├── users/
│   │       ├── objects/
│   │       └── .../
│   │
│   ├── domain/                  # Business logic (no dependencies)
│   │   ├── auth/
│   │   │   ├── entity.go       # User, Session entities
│   │   │   └── service.go      # AuthService
│   │   ├── users/
│   │   ├── objects/
│   │   └── .../
│   │
│   ├── infrastructure/
│   │   ├── db/
│   │   │   └── context.go     # RunInTenantTx, RunInTx helpers
│   │   │
│   │   └── middleware/
│   │       ├── tenant.go       # ExtractTenant middleware
│   │       └── auth.go        # JWT validation middleware
│   │
│   └── config/
│       └── config.go           # Configuration structs
│
├── pkg/
│   └── errors/
│       └── errors.go          # AppError types
│
├── migrations/
│   ├── global/                  # Global schema (eamtenants)
│   └── tenant/                 # Tenant schema
│
├── deploy/
│   └── pgbouncer/             # PgBouncer config files
│
├── go.mod
├── go.sum
├── Makefile
├── config.yaml
└── .gitignore
```

### Layer Responsibilities

| Layer              | Files                         | Responsibility                         |
| ------------------ | ----------------------------- | -------------------------------------- |
| **API**            | `adapters/api/*/handler.go`   | HTTP request/response, parse, validate |
| **Domain**         | `domain/*/service.go`         | Business logic, orchestration          |
| **Repository**     | `adapters/db/*/repository.go` | Data access, SQL execution             |
| **Infrastructure** | `infrastructure/`             | Cross-cutting (DB, middleware)         |
| **Config**         | `config.go`, `config.yaml`    | Configuration management               |

---

## Key Files Reference

### `infrastructure/db/context.go`

```go
// RunInTenantTx executes a function within a tenant-scoped transaction.
// It automatically sets the search_path for the transaction.
func RunInTenantTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error

// RunInTx executes a function within a transaction without tenant scope.
// Use for global operations (e.g., tenant creation).
func RunInTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error
```

### `infrastructure/middleware/tenant.go`

```go
// ExtractTenant middleware extracts tenant from query/header/body
// and stores it in the request context.
func (m *TenantMiddleware) ExtractTenant() fiber.Handler

// GetTenant retrieves the tenant code from context
func GetTenant(c *fiber.Ctx) string
```

### `adapters/db/*/repository.go` (example)

```go
func (r *PgUserRepository) FindByCode(ctx context.Context, code string) (*User, error) {
    var user *User
    err := infraDB.RunInTenantTx(ctx, r.pool, func(tx pgx.Tx) error {
        return tx.QueryRow(ctx, "SELECT * FROM eamusers WHERE usr_code = $1", code).Scan(...)
    })
    return user, err
}
```

---

## Environment Variables

| Variable          | Default                 | Description                          |
| ----------------- | ----------------------- | ------------------------------------ |
| `DB_HOST`         | `localhost`             | PostgreSQL host                      |
| `DB_PORT`         | `5432`                  | PostgreSQL port                      |
| `DB_USER`         | `postgres`              | Database user                        |
| `DB_PASSWORD`     | `postgres`              | Database password                    |
| `DB_DATABASE`     | `nova`                  | Database name                        |
| `DB_SCHEMA`       | `public`                | Default schema                       |
| `JWT_SECRET`      | `your-super-secret-key` | JWT signing secret                   |
| `JWT_EXPIRY_MINS` | `15`                    | JWT token expiry in minutes          |
| `SERVER_PORT`     | `4000`                  | HTTP server port                     |
| `SERVER_ENV`      | `development`           | Environment (development/production) |

---

## Testing

```bash
# Run all tests
make test

# Run with coverage
go test -v -cover ./...

# Run specific package
go test -v ./internal/adapters/db/auth/...
```

---

## Troubleshooting

### "Tenant connection not found"

```
[WARN] Tenant 'acme' set but no tenant-scoped connection available.
Using pool without tenant isolation.
```

**Cause:** Middleware not running, or tenant not provided in request.

**Fix:** Ensure `ExtractTenant` middleware is registered and request includes `?tenant=acme`.

### "relation does not exist"

```
ERROR: relation "eamusers" does not exist
```

**Cause:** Wrong schema. Either:

1. Tenant not set in request
2. Tenant schema not created
3. search_path not set correctly

**Fix:**

- Check request includes `?tenant=acme`
- Verify schema exists: `SELECT schemaname FROM pg_catalog.pg_tables WHERE tablename = 'eamusers'`

### Slow queries with many tenants

**Cause:** No index on `usr_tenant_id` in tenant schema tables.

**Fix:** Add indexes:

```sql
CREATE INDEX idx_eamusers_tenant ON eamusers(usr_tenant_id);
CREATE INDEX idx_eamobjects_tenant ON eamobjects(obj_tenant_id);
-- etc.
```
