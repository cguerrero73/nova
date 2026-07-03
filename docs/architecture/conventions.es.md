# Nova — Convenciones de Arquitectura

> Fuente de verdad para decisiones arquitectónicas transversales.
> Todos los módulos nuevos DEBEN seguir estas convenciones.
> Versión en inglés: [`conventions.md`](conventions.md)

---

## 1. Multi-Tenancy

**Patrón:** Schema-per-tenant en un cluster PostgreSQL compartido.

- Cada tenant tiene su propio esquema de PostgreSQL (ej: `tenant_acme`, `tenant_xyz`).
- El esquema `public` contiene datos cross-tenant (`eamtenants`, etc.).
- El middleware `ExtractTenant` extrae el código de tenant del request:
  1. Query param `?tenant=<code>`
  2. Header `X-Tenant-Code`
  3. Body JSON `{"tenant": "<code>"}`
- `RunInTenantTx(ctx, pool, fn)` ejecuta `SET search_path TO tenant_<code>, public` al inicio de la transacción.
- Todas las queries sobre tablas con scope de tenant DEBEN pasar por `RunInTenantTx`.

**Reglas para módulos nuevos:**

- NUNCA agregar columnas `tenant_id` a tablas nuevas.
- NUNCA agregar predicados `tenant_id` a las queries.
- Las tablas viven en el schema del tenant — el aislamiento es trabajo del middleware, no del schema.
- Las restricciones de unicidad son per-schema por construcción.

**Archivos fuente:**

- `backend/internal/infrastructure/middleware/tenant.go`
- `backend/internal/infrastructure/db/context.go`

---

## 2. Autenticación (JWT)

**JWT = solo identidad.** El token porta quién es el usuario, NO qué puede hacer.

**Claims del token:**

| Claim | Propósito |
|---|---|
| `userCode` | Identificador del usuario dentro del tenant |
| `email` | Email del usuario |
| `name` | Nombre para mostrar |
| `tenant` | Código de tenant (pista para routing de schema) |

**Cómo acceder en los handlers:**

```go
claims := middleware.GetUserClaims(c)
userCode := claims.UserCode
tenant   := claims.Tenant
```

**Cadena de middlewares:**

1. `ExtractTenant` → establece `search_path` en la conexión DB
2. `Authenticate` → valida JWT, guarda claims en `c.Locals("user")`
3. `ContextLoader` → lee el rol activo de `eamsessions`, lo guarda en `c.Locals("activeRole")`

**Archivos fuente:**

- `backend/internal/domain/auth/entity.go` — struct `TokenClaims`
- `backend/internal/infrastructure/middleware/auth.go` — `AuthMiddleware`, `GetUserClaims`

---

## 3. Roles

**Un rol activo por usuario a la vez.** Un usuario puede tener múltiples roles en distintas organizaciones, pero solo uno está activo en una sesión dada.

**Schema:**

| Tabla | Propósito |
|---|---|
| `eamroles` | Definiciones de roles por tenant (código, descripción, flag system). Seedeados: `ADMIN`, `EMPTY` |
| `eamuser_organizations` | Usuarios ↔ organizaciones ↔ roles. `uog_default` marca la org+rol default del usuario |
| `eamsessions.ses_active_role` | Rol actualmente activo para la sesión |

**Resolución del rol activo:**

1. En el login: leer el rol default del usuario (de `eamuser_organizations` donde `uog_default = '+'`) y escribirlo en `eamsessions.ses_active_role`.
2. En cada request: el middleware `ContextLoader` lee `ses_active_role` de la fila de sesión y lo guarda en `c.Locals("activeRole")`.
3. Cambio de rol: `POST /api/auth/switch-context { "role": "<role_code>" }` → valida que el usuario tiene ese rol → `UPDATE eamsessions SET ses_active_role` → el frontend recarga el menú.

**El JWT NO se modifica al cambiar de rol.** El JWT sigue siendo válido para identidad; el contexto de autorización cambia via la fila de sesión.

**Reglas para módulos nuevos:**

- Siempre leer el rol activo de `c.Locals("activeRole")`, nunca de los claims del JWT.
- El cambio de rol requiere recargar el menú completo en el frontend (los permisos cambian).

---

## 4. Organizaciones

**Un usuario tiene acceso a N organizaciones simultáneamente.** Una es la default, usada para autocompletar el campo de organización al crear registros en entidades multi-org.

**Schema:**

| Tabla | Propósito |
|---|---|
| `eamorganizations` | Definiciones de organizaciones (código, nombre, flag common) |
| `eamusers.usr_default_org` | Organización default del usuario para autocompletar |

**Reglas para módulos nuevos:**

- La organización NO es parte del contexto de sesión.
- Para entidades con scope por organización (ej: parts, stores), el campo org es llenado por el frontend usando la org default del usuario.
- El acceso multi-org significa que un usuario puede consultar datos de distintas orgs en la misma sesión.

---

## 5. Permisos

**Patrón:** `HasPermission(screen, action)` genérico — NO las columnas CRUD fijas legacy.

**Modelo target** (normalizado):

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

**Estado actual** (deuda técnica — migración pendiente):

El `eamrole_permissions` actual tiene columnas fijas: `rpe_select`, `rpe_insert`, `rpe_update`, `rpe_delete`, `rpe_print`. Estas deben migrarse al modelo normalizado `(screen, action, allowed)`.

**Cómo funciona:**

```go
role.HasPermission("formbuilder", "publish")  // true / false
role.HasPermission("parts", "select")          // true / false
```

Wildcards soportados: screen `*` o acción `*` otorga acceso amplio.

**Reglas para módulos nuevos:**

- Definir **acciones semánticas**, no mapeos CRUD. Ejemplos: `formbuilder.view`, `formbuilder.design`, `formbuilder.publish`, `formbuilder.assign`.
- NO forzar acciones de módulos nuevos a `select/insert/update/delete`. Si la acción no es CRUD, nombrarla por lo que hace.
- El chequeo de permisos va en el handler, después de auth:

```go
claims := middleware.GetUserClaims(c)
role   := c.Locals("activeRole").(string)
// cargar role de DB, luego:
if !role.HasPermission("formbuilder", "design") {
    return c.Status(403).JSON(...)
}
```

**Archivos fuente:**

- `backend/internal/domain/roles/entity.go` — `Role.HasPermission(screen, action)` (ya implementado, esperando uso)

---

## 6. Cadena de Middlewares (orden esperado)

```
Request
  → ExtractTenant        (establece search_path)
  → Authenticate         (valida JWT → c.Locals("user"))
  → ContextLoader        (lee eamsessions.ses_active_role → c.Locals("activeRole"))
  → [route handler]      (lee locals, verifica permisos, ejecuta lógica de negocio)
```

---

## 7. Cheat Sheet para Módulos Nuevos

| Pregunta | Respuesta |
|---|---|
| ¿De dónde viene el tenant? | `search_path` seteado por `ExtractTenant`. Nunca leer `tenant_id`. |
| ¿De dónde viene el rol activo? | `c.Locals("activeRole")`, cargado de `eamsessions` por `ContextLoader`. |
| ¿Cómo verifico permisos? | `role.HasPermission(screen, action)` después de cargar el rol. |
| ¿Debo agregar `tenant_id` a mi tabla? | **No.** Nunca. El schema aísla. |
| ¿Debo poner el rol en el JWT? | **No.** JWT es solo identidad. El rol vive en la sesión. |
| ¿Qué acciones defino para mi módulo? | Acciones semánticas nombradas por lo que hacen, no CRUD. |
| ¿Cómo manejo multi-org? | Frontend usa `usr_default_org` para autocompletar. Org no es contexto de sesión. |
