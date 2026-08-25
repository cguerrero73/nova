# Nova — Architecture Conventions

> Source of truth for cross-cutting architectural decisions.
> All new modules MUST follow these conventions.
> Spanish version: [`conventions.es.md`](conventions.es.md)

---

## 1. Multi-Tenancy

**Pattern:** Schema-per-tenant on a shared PostgreSQL cluster.

- Each tenant has its own PostgreSQL schema (e.g. `tenant_acme`, `tenant_xyz`).
- The `public` schema holds cross-tenant data (`eamtenants`, etc.).
- `ExtractTenant` middleware extracts the tenant code from the request:
  1. Query param `?tenant=<code>`
  2. Header `X-Tenant-Code`
  3. Body JSON `{"tenant": "<code>"}`
- `RunInTenantTx(ctx, pool, fn)` executes `SET search_path TO tenant_<code>, public` at transaction start.
- All queries on tenant-scoped tables MUST go through `RunInTenantTx`.

**Rules for new modules:**

- NEVER add `tenant_id` columns to new tables.
- NEVER add `tenant_id` predicates to queries.
- Tables live in the tenant's schema — isolation is the middleware's job, not the schema's job.
- Uniqueness constraints are per-schema by construction.

**Source files:**

- `backend/internal/infrastructure/middleware/tenant.go`
- `backend/internal/infrastructure/db/context.go`

---

## 2. Authentication (JWT)

**JWT = identity only.** The token carries who the user is, NOT what they can do.

**Token claims:**

| Claim | Purpose |
|---|---|
| `userCode` | User identifier within the tenant |
| `email` | User email |
| `name` | Display name |
| `tenant` | Tenant code (schema routing hint) |

**How to access in handlers:**

```go
claims := middleware.GetUserClaims(c)
userCode := claims.UserCode
tenant   := claims.Tenant
```

**Middleware chain:**

1. `ExtractTenant` → sets `search_path` on DB connection
2. `Authenticate` → validates JWT, stores claims in `c.Locals("user")`
3. `ContextLoader` → reads active role from `eamsessions`, stores in `c.Locals("activeRole")`

**Source files:**

- `backend/internal/domain/auth/entity.go` — `TokenClaims` struct
- `backend/internal/infrastructure/middleware/auth.go` — `AuthMiddleware`, `GetUserClaims`

---

## 3. Roles

**One active role per user at a time.** A user may hold multiple roles across organizations, but only one is active in a given session.

**Schema:**

| Table | Purpose |
|---|---|
| `eamroles` | Role definitions per tenant (code, description, system flag). Seeded: `ADMIN`, `EMPTY` |
| `eamuser_organizations` | Users ↔ organizations ↔ roles. `uog_default` marks the user's default org+role |
| `eamsessions.ses_active_role` | Currently active role for the session |

**Active role resolution:**

1. On login: read the user's default role (from `eamuser_organizations` where `uog_default = '+'`) and write it to `eamsessions.ses_active_role`.
2. On each request: `ContextLoader` middleware reads `ses_active_role` from the session row and stores it in `c.Locals("activeRole")`.
3. Role switch: `POST /api/auth/switch-context { "role": "<role_code>" }` → validates the user holds that role → `UPDATE eamsessions SET ses_active_role` → frontend reloads menu.

**JWT is NOT modified on role switch.** The JWT remains valid for identity; the authorization context changes via the session row.

**Rules for new modules:**

- Always read the active role from `c.Locals("activeRole")`, never from JWT claims.
- Role switch requires a full menu reload on the frontend (permissions change).

---

## 4. Organizations

**A user has access to N organizations simultaneously.** One is the default, used to auto-complete the organization field when creating records in multi-org entities.

**Schema:**

| Table | Purpose |
|---|---|
| `eamorganizations` | Organization definitions (code, name, common flag) |
| `eamusers.usr_default_org` | User's default organization for auto-complete |

**Rules for new modules:**

- Organization is NOT part of the session context.
- For entities scoped by organization (e.g. parts, stores), the org field is filled by the frontend using the user's default org.
- Multi-org access means a user can query data from different orgs in the same session.

---

## 5. Permissions

**Pattern:** Generic `HasPermission(screen, action)` — NOT the legacy fixed CRUD columns.

**Target model** (normalized):

```sql
eamrole_permissions (
    rpe_id      BIGSERIAL PRIMARY KEY,
    rpe_role    VARCHAR(50)  NOT NULL REFERENCES eamroles(rol_code),
    rpe_screen  VARCHAR(100) NOT NULL,
    rpe_action  VARCHAR(50)  NOT NULL,
    rpe_allowed BOOLEAN      NOT NULL DEFAULT false,
    ...
    CONSTRAINT uq_rpe UNIQUE (rpe_role, rpe_screen, rpe_action)
)
```

**Legacy state** (debt — pending migration):

Current `eamrole_permissions` has fixed columns: `rpe_select`, `rpe_insert`, `rpe_update`, `rpe_delete`, `rpe_print`. These must be migrated to the normalized `(screen, action, allowed)` model.

**How it works:**

```go
role.HasPermission("formbuilder", "publish")  // true / false
role.HasPermission("parts", "select")          // true / false
```

Wildcards supported: `*` screen or `*` action grants broad access.

**Rules for new modules:**

- Define **semantic actions**, not CRUD mappings. Examples: `formbuilder.view`, `formbuilder.design`, `formbuilder.publish`, `formbuilder.assign`.
- Do NOT force new module actions into `select/insert/update/delete`. If the action isn't CRUD, name it for what it does.
- Permission check goes in the handler, after auth:

```go
claims := middleware.GetUserClaims(c)
role   := c.Locals("activeRole").(string)
// load role from DB, then:
if !role.HasPermission("formbuilder", "design") {
    return c.Status(403).JSON(...)
}
```

**Source files:**

- `backend/internal/domain/roles/entity.go` — `Role.HasPermission(screen, action)` (already implemented, awaiting use)

---

## 6. Middleware Chain (expected order)

```
Request
  → ExtractTenant        (sets search_path)
  → Authenticate         (validates JWT → c.Locals("user"))
  → ContextLoader        (reads eamsessions.ses_active_role → c.Locals("activeRole"))
  → [route handler]      (reads locals, checks permissions, runs business logic)
```

---

## 7. Cheat Sheet for New Modules

| Question | Answer |
|---|---|
| Where does tenant come from? | `search_path` set by `ExtractTenant`. Never read `tenant_id`. |
| Where does the active role come from? | `c.Locals("activeRole")`, loaded from `eamsessions` by `ContextLoader`. |
| How do I check permissions? | `role.HasPermission(screen, action)` after loading the role. |
| Should I add `tenant_id` to my table? | **No.** Ever. The schema isolates. |
| Should I put the role in the JWT? | **No.** JWT is identity only. Role lives in the session. |
| What actions do I define for my module? | Semantic actions named after what they do, not CRUD. |
| How do I handle multi-org? | Frontend uses `usr_default_org` for auto-complete. Org is not a session context. |
