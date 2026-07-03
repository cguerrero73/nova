# Dynamic Form Builder — Architecture Proposal

> **Status:** Pre-SDD proposal for human review.
> **Audience:** Nova core team.
> **Goal:** Lock architectural decisions BEFORE generating formal SDD artifacts (`openspec/{proposal,spec,design,tasks}`).

---

## 1. Executive Summary

Nova will gain a **Dynamic Form Builder** module: a versioned, role-aware form definition system with a visual designer for privileged users and a runtime renderer for end users, scoped per tenant via Nova's existing connection/schema isolation (see Section 2.0). A **Form** is a logical entity; under each form there are N independent **Layouts** (one per role audience), each of which is itself a complete, versioned JSON document. Versioning is **per-layout, not per-form** — publishing a new version of "admin-full" never touches "agent-compact". Every form automatically ships a reserved layout named **`default`**, which acts as the implicit fallback for any role without a specific assignment: the backend resolves `(form_key, role) → assignment → layout → published_version` against the tenant's own database/schema (which the existing request middleware has already resolved), and if no assignment exists for the caller's role it falls back to the `default` layout instead of erroring out. The frontend renders the resolved JSON via Angular Reactive Forms, and the designer uses Angular CDK DragDrop for reorder/move operations between sections.

The three most consequential architectural decisions are:

1. **Versioning is per-layout, not per-form, with immutable published snapshots.** Each tenant owns a set of named layouts per form (e.g., `admin-full`, `agent-compact`, `viewer-readonly`); each layout has its own stack of append-only `published` versions and at most one mutable `draft`. Publishing is a snapshot operation, never an in-place update. This gives us rollback, audit, and per-role review without the complexity of optimistic locking, and lets one role's evolution proceed independently of another's.
2. **Layouts are role-keyed, not role-filtered, with a `default` fallback for unassigned roles.** Each role audience has its **own complete layout JSON**, not a filtered view of a master definition. Resolution is `(form_key, role) → role_layout_assignments → layout → current_published_version`, with no filtering happening at runtime. Tenant scoping is handled implicitly by the connection/schema the request arrived on (Section 2.0) — it is NOT part of the resolution tuple. If no active assignment exists for the caller's role, the resolver falls back to the layout named `default` for that form — every form ships one, and it represents "what any user sees if we haven't customized their role's view." The `default` name is reserved within a form: it is auto-created on form creation, cannot be renamed or deleted while the form exists (it can be edited in place), and the service layer rejects attempts to create a second layout with that name. The designer is the only place that decides what each role sees; the backend trusts the assignment table, and trusts the `default` layout as the catch-all.
3. **The form definition JSON Schema is the single source of truth, validated by Zod on both sides.** Schemas live in a new top-level `shared/` directory consumed by the Go API (via `go:embed` of the JSON Schema) and the Angular app. Any drift is a build-time error, not a runtime crash.

---

## 2. Domain Model

### 2.0 Multi-tenancy model (read this first)

Nova's multi-tenancy is **schema-per-tenant on a shared PostgreSQL cluster**, not row-level. Concretely:

- Each tenant has its own PostgreSQL schema (e.g. `tenant_acme`, `tenant_xyz`). The `public` schema holds cross-tenant data (`eamtenants`, etc.).
- The existing `ExtractTenant` middleware reads the tenant code from the request (query param, header, or body), and the `RunInTenantTx` helper sets `SET search_path TO tenant_<code>, public` at the start of every transaction.
- Therefore, when a request handler calls `RunInTenantTx(ctx, pool, fn)`, every SQL inside `fn` runs against the **caller's tenant schema only**. There is no `tenant_id` predicate on any query and no shared table that mixes rows across tenants.

**Implication for this module.** The form builder is a *new* module. It is **not** legacy, so it does not inherit the `*_tenant_id` columns that some older tables carry for historical reasons. Every table in this module (`eamform_definitions`, `eamform_layouts`, `eamform_layout_versions`, `eamform_role_assignments`, `eamform_audit_log`, `eamform_submissions`) lives in the tenant's schema and **carries no `tenant_id` column**. Isolation is the middleware's job, not the schema's job and not the application's job.

**Cross-tenant implications:**

- A `form_key` (e.g. `customer-intake`) is unique **within a single tenant's schema only**. Two tenants can independently define a form called `customer-intake` — they cannot collide because the data physically lives in different schemas.
- The `actor_user_id` in the audit log is the user *within the tenant* that owns the schema the audit row is in. It is not a global user ID. There is no cross-tenant user reference.
- Cache keys, error logs, observability traces **may** include `tenant_id` for human debugging (it's just a label), but the database layer never uses it for isolation.

**What the form builder module does NOT add:**

- No `tenant_id` columns on any table.
- No `tenant_id` predicate in any query (the connection's `search_path` already restricts rows).
- No tenant-aware middleware, connection management, or scoping logic. The existing `RunInTenantTx` / `SetTenantSchema` helpers are used unchanged.
- No cross-tenant lookups anywhere. A form is always served from the caller's own schema.

This is simpler and stricter than the alternative (shared schema with `tenant_id` column + row-level filtering) and matches the user's instruction: *"la grabación del layout ya va dirigida al tenant, eso es por conexión y se grabará en distintas bd, ya se resuelve por default en la gestión de los request, no es necesario especificarlo como dato."*

### 2.1 Why "layouts are role-keyed, not role-filtered"

The previous design stored one form definition per form and applied per-role visibility/required/readonly policy at runtime by filtering fields. That has three structural problems:

1. **Schema drift between audiences.** Admin, agent, and viewer often need *different fields*, *different layouts*, and *different validation* — not just "some fields hidden". Filtering one document can produce malformed layouts (e.g., a section rendered with zero fields, or a required address whose country is hidden).
2. **Review surface is muddy.** A change to "what the agent sees" appears as a tiny diff inside an otherwise-unrelated admin form. Reviewers cannot reason about one role's UX without reading the whole definition.
3. **Runtime cost on the hot path.** Every read parses the master JSON and walks the policy tree. The cost is unbounded as fields grow.

The new model makes each role's UX a **first-class, independent artifact**: a named `layout` with its own complete JSON document and its own version history. Resolution is a join, not a filter.

### 2.2 ER diagram

No table in this module carries a `tenant_id` column — see Section 2.0. Every row is implicitly scoped to the tenant whose schema the connection's `search_path` points to.

```
erDiagram
    eamform_definitions ||--o{ eamform_layouts : "has many layouts"
    eamform_definitions ||--o{ eamform_audit_log : "audited"

    eamform_definitions {
        bigint frm_id PK
        text    frm_key          "logical key, unique within the tenant's schema"
        text    frm_name
        text    frm_description
        text    frm_status "active | archived"
        text    frm_created_by
        timestamptz frm_created_at
        timestamptz frm_updated_at
    }

    eamform_layouts ||--o{ eamform_layout_versions : "versioned snapshots"
    eamform_layouts ||--o{ eamform_role_assignments : "mapped from roles"
    eamform_layouts {
        bigint  fl_id PK
        bigint  fl_form_id FK
        text    fl_name "slug, unique within (form)"
        text    fl_display_name
        text    fl_description
        text    fl_status "active | archived"
        bigint  fl_draft_version_id FK "nullable pointer to current draft"
        bigint  fl_published_version_id FK "nullable pointer to currently published"
        text    fl_created_by
        timestamptz fl_created_at
        timestamptz fl_updated_at
        timestamptz fl_archived_at "nullable"
    }

    eamform_layout_versions {
        bigint  flv_id PK
        bigint  flv_layout_id FK
        int     flv_version_number "monotonic per layout"
        text    flv_kind "draft | published | archived"
        text    flv_description "change note (commit-message style)"
        jsonb   flv_definition "the complete layout JSON (see Section 3)"
        text    flv_created_by
        timestamptz flv_created_at
        timestamptz flv_published_at "nullable"
    }

    eamform_role_assignments {
        bigint  fra_id PK
        bigint  fra_form_id FK
        bigint  fra_layout_id FK
        text    fra_role_name "e.g. admin, agent, viewer"
        text    fra_assigned_by
        timestamptz fra_assigned_at
        timestamptz fra_revoked_at "nullable"
    }

    eamform_audit_log {
        bigint  fal_id PK
        text    fal_actor_user_id "user within the tenant who did it"
        text    fal_action "form.create | layout.create | layout.publish | ..."
        text    fal_entity_type "form | layout | assignment | version"
        bigint  fal_entity_id
        jsonb   fal_metadata "action-specific payload (from/to versions, diff summary)"
        text    fal_note "optional human note"
        timestamptz fal_created_at
    }

    eamform_submissions {
        bigint  fs_id PK
        bigint  fs_form_id FK
        bigint  fs_layout_id FK "which layout served the form"
        bigint  fs_version_id FK "snapshot of the version that accepted it"
        text    fs_actor_id
        jsonb   fs_payload "submitted data"
        text    fs_status "submitted | rejected"
        timestamptz fs_created_at
    }
```

### 2.3 Tables

#### `eamform_definitions`

The stable identity of a logical form *within the tenant's schema*. Holds metadata only — the form has no "current version" pointer anymore; that lives on each layout.

| Column | Purpose |
|---|---|
| `frm_id` | Surrogate PK. |
| `frm_key` | Stable string key within the tenant's schema (e.g. `customer-intake`). `UNIQUE (frm_key)` — uniqueness is per-schema by construction, since each tenant has its own schema. Used by URL paths and code references — survives layout churn. |
| `frm_name`, `frm_description` | Display + admin metadata. |
| `frm_status` | `active` or `archived`. Archived forms are not editable and not served at runtime, but their layouts and submissions remain queryable. |
| `frm_created_by`, `frm_created_at`, `frm_updated_at` | Audit trail (light — full trail is in `eamform_audit_log`). |

> **Why no `frm_tenant_id`?** The tenant is implicit via `search_path`. A second tenant may have its own `customer-intake` row in a different schema; the schemas are physically isolated, so no collision and no predicate needed. See Section 2.0.

#### `eamform_layouts`

One row per *role audience* for a form. Identified by `(form_id, name)` within the tenant's schema.

| Column | Purpose |
|---|---|
| `fl_id` | Surrogate PK. |
| `fl_form_id` | FK to `eamform_definitions`. |
| `fl_name` | Slug-friendly identifier, unique per `fl_form_id`. Examples: `admin-full`, `agent-compact`, `viewer-readonly`, `manager-overview`. Lowercase, dashes, alphanumeric. The literal value `default` is **reserved**: every form auto-ships one layout with this name (see "Reserved name: `default`" below). |
| `fl_display_name` | Human-friendly title shown in the designer UI. |
| `fl_description` | One-line purpose, shown in the layout picker. |
| `fl_status` | `active` or `archived`. Archived layouts are not served at runtime but their version history is preserved. |
| `fl_draft_version_id` | Pointer to the current working copy in `eamform_layout_versions`. `NULL` means no draft yet. |
| `fl_published_version_id` | Pointer to the version runtime serves. `NULL` means never published. |
| `fl_created_by`, `fl_created_at`, `fl_updated_at`, `fl_archived_at` | Audit trail. |

**Unique constraint:** `UNIQUE (fl_form_id, fl_name)`. This is what the designer UI uses to identify layouts and what the backend uses for assignments. Uniqueness is per-schema by construction (Section 2.0).

**Is `fl_name` mutable?** No. Rename is achieved by creating a new layout, migrating assignments to the new one, and archiving the old. Justification: assignments, URLs, and audit entries reference the layout by ID (and historically by name); renames would orphan those references and break human-readable audit trails. Create-new-and-archive-old is also what `publish` does to versions, so the pattern is consistent.

**Reserved name: `default`.** Within every form, the literal value `"default"` for `fl_name` carries special meaning: it is the implicit fallback layout that the resolver returns when a role has no active assignment in `eamform_role_assignments` for that form. Concretely:

- Every form gets a layout named `default` **automatically created** at the moment the form row is inserted. The same transaction that writes `eamform_definitions` writes the initial `eamform_layouts` row for `default` (empty draft, no published version yet) and the corresponding audit entry.
- The reserved name is enforced **at the service layer**, not by a DB `CHECK` constraint. Justification: a `CHECK (fl_name <> 'default')` would also reject the very first auto-creation; a partial constraint `CHECK (fl_name <> 'default' OR some_marker_column)` couples the name to schema-level state we don't want to encode. The application invariant — "at most one layout named `default` per form, and exactly one such layout must exist for every non-archived form" — is what we actually want, and it is best expressed in code where we can also write the audit entry for the rejection. The DB still enforces **uniqueness** via `UNIQUE (fl_form_id, fl_name)`; the reserved-name rule is the layer on top.
- The service layer rejects: (i) `CreateLayout` when the requested name is `default` and one already exists for the form; (ii) attempts to rename a layout to `default` (which today is impossible because renames are forbidden, but the rule still holds — no future "rename support" can sneak it past); (iii) `ArchiveLayout` targeting `default` while the form is still `active` (the form must first be archived, after which the `default` layout can be cleaned up — but in practice we cascade-archive the form and never expose an explicit `default` delete path).
- The DB additionally enforces "every form has at least one `default` layout" at deletion time: a trigger on `eamform_definitions` `BEFORE DELETE` raises if any row in `eamform_layouts` still references this form with `fl_name = 'default'`. Rationale: cascading the layout delete at the FK level would orphan submissions that referenced the `default` version, and submissions are immutable. The trigger ensures an admin must explicitly purge layouts + submissions before deleting a form. Archiving (soft-delete) of the form is allowed and does not require touching the `default` layout — archived forms are simply not served at runtime.

#### `eamform_layout_versions`

Append-only history of a layout. **Never updated after insert.** Publishing creates a *new* row, never mutates the old one.

| Column | Purpose |
|---|---|
| `flv_id` | Surrogate PK. |
| `flv_layout_id` | FK to `eamform_layouts`. |
| `flv_version_number` | Monotonic integer per layout (`1, 2, 3…`). Assigned by the application at insert time (max + 1 in transaction). |
| `flv_kind` | `draft`, `published`, `archived`. Partial unique index `(flv_layout_id) WHERE flv_kind = 'draft'` enforces one-draft-per-layout. |
| `flv_description` | Designer-supplied change note (like a git commit message). Filled by the human publishing the version — see the note on `description` at the end of this section. |
| `flv_definition` | `jsonb` — the **complete** layout JSON (see Section 3). Validated by Zod on write. |
| `flv_created_by`, `flv_created_at`, `flv_published_at` | Audit trail. |

**Immutability is enforced by trigger**: `BEFORE UPDATE` on `eamform_layout_versions` raises an exception. Drafts become "editable" by inserting a new draft row and repointing `fl_draft_version_id`; the prior draft row is left in place (and is later discarded if the user wants to start over, but the audit trail is preserved).

#### `eamform_role_assignments`

Many-to-many between forms and roles via layouts. Each row says "for this form, this role serves this layout".

| Column | Purpose |
|---|---|
| `fra_id` | Surrogate PK. |
| `fra_form_id` | FK to `eamform_definitions`. |
| `fra_layout_id` | FK to `eamform_layouts`. |
| `fra_role_name` | The role key (`admin`, `agent`, etc.). Roles are tenant-internal — they live in the tenant's user-management tables and are not global across tenants. |
| `fra_assigned_by`, `fra_assigned_at`, `fra_revoked_at` | Audit trail. `revoked_at` non-null is a soft-delete; we never hard-delete because submissions remain tied to the historical assignment. |

**Constraints:**
- Partial unique `(fra_form_id, fra_role_name) WHERE fra_revoked_at IS NULL` — at most one **active** assignment per `(form, role)`.
- `(fra_form_id, fra_role_name, fra_assigned_at DESC)` index for "what does this role see right now?" lookups.
- The assigned `fra_layout_id` must belong to the same `fra_form_id` — enforced by trigger or `CHECK` join.

A role can have zero active assignments for a given form. Runtime behavior in that case is **not** an error: the resolver falls back to the layout named `default` for that form (see Section 4 and the "Reserved name: `default`" note on `eamform_layouts` above). Designers can also explicitly assign the `default` layout to a role via the assignment UI — this is redundant but legal, and is sometimes useful as a self-documenting signal in the assignments table. The system never returns `FormLayoutNotAssigned`; missing-assignment always resolves to `default`.

#### `eamform_audit_log`

Append-only, immutable record of every mutation. Same write-once model as `eamform_layout_versions`.

| Column | Purpose |
|---|---|
| `fal_id` | Surrogate PK (bigint, monotonic). |
| `fal_actor_user_id` | **Who** did it. Always populated. This is the user ID within the tenant whose schema the audit row lives in — it is NOT a global user identifier. The tenant is implicit via `search_path` (Section 2.0). |
| `fal_action` | **What** they did. One of an enum (see below). |
| `fal_entity_type` | **Which kind** of thing: `form`, `layout`, `assignment`, `version`. |
| `fal_entity_id` | **Which instance** — ID into the corresponding table. |
| `fal_metadata` | `jsonb` — action-specific payload. Examples: `{ "from_version": 7, "to_version": 8 }` for `layout.publish`, `{ "layout_name": "agent-compact" }` for `layout.create`. |
| `fal_note` | Optional human note (designer comment, reason for archive). |
| `fal_created_at` | **When**. `timestamptz`, default `now()`. |

**Action enum (v1):**

```
form.create, form.archive,
layout.create, layout.archive, layout.assign, layout.unassign,
version.draft_save, version.publish, version.revert
```

**Immutability:** enforced by trigger (`BEFORE UPDATE` and `BEFORE DELETE` both raise). The only writer is the service layer — no application code ever mutates or deletes an audit row.

**Automatic vs explicit?** **Explicit** (service-layer writes). Rationale:

- Triggers see row-level context but not the *intent* (the note, the metadata about a role assignment's old/new state). The service layer has the full picture.
- Triggers can't easily attribute actions to a user — the user comes from the auth context, which lives in the application.
- Service-layer writes are testable: we can assert "creating a layout also writes the audit row" in the same unit test.
- Cost is one extra `INSERT` per mutation — negligible.

**Retention:** indefinite by default. Layout versions, audit, and submissions stay forever; archival hides them from the UI but they remain queryable for compliance. A future GDPR-style retention job can purge rows per schema (i.e., per tenant) and entity.

#### `eamform_submissions`

Captures submitted data alongside the exact layout version that accepted it. Out of scope for v1 but referenced because `flv_id` and `fl_id` must exist in the schema from day one.

### 2.4 Concrete example: "customer-intake" with 3 layouts

All rows below live in the request's tenant schema (e.g. `tenant_acme`). The tenant code does not appear in the row — `search_path` already routed the query here.

```
eamform_definitions
─────────────────────
frm_id: 42
frm_key: "customer-intake"
frm_name: "Customer Intake"
frm_description: "Initial data capture for new customers."

eamform_layouts (3 rows for this form)
──────────────────────────────────────
fl_id | fl_form_id | fl_name          | fl_draft_version_id | fl_published_version_id
-----+-----------+------------------+---------------------+------------------------
100  | 42        | default          | 1010                | 1009
101  | 42        | admin-full       | 1004                | 1003
102  | 42        | agent-compact    | 1006                | 1005

eamform_role_assignments (2 rows for this form)
───────────────────────────────────────────────
fra_role_name | fra_layout_id | fra_revoked_at
--------------+---------------+---------------
admin         | 101           | NULL
agent         | 102           | NULL

eamform_layout_versions (subset)
───────────────────────────────
flv_id | flv_layout_id | flv_version_number | flv_kind      | flv_description
------+---------------+--------------------+---------------+-------------------------
1001  | 101           | 1                  | archived      | initial draft
1002  | 101           | 2                  | published     | first admin-full
1003  | 101           | 3                  | published     | added tax_id section
1004  | 101           | 4                  | draft         | WIP: split consent section
1005  | 102           | 1                  | published     | first agent-compact
1006  | 102           | 2                  | draft         | WIP: hide consent for agents
1008  | 100           | 1                  | archived      | initial default skeleton
1009  | 100           | 2                  | published     | baseline form for unassigned roles
1010  | 100           | 3                  | draft         | WIP: tighten required fields
```

Note the absent rows for `viewer`, `manager`, and `auditor` in `eamform_role_assignments`. There is **no** `viewer-readonly` layout anymore — that case is now served by `default`. See Section 4.1 for the runtime trace.

Resolution for an `agent` user requesting form `customer-intake`:

1. The middleware has already opened a connection with `search_path = tenant_acme, public`. The handler reads the request context and proceeds.
2. Look up active assignment: `fra_role_name = 'agent'` for `frm_key='customer-intake'` → `fra_layout_id = 102`.
3. Look up `fl_id = 102` → `fl_published_version_id = 1005`.
4. Return `flv_definition` of version 1005. That's it.

For a `viewer` (no row in `eamform_role_assignments`): the same flow runs step 2 and finds nothing, then steps 3–4 run against the `default` layout (`fl_id = 100`) → `flv_id = 1009`.

No filtering. No policy walk. No JSON traversal. No `FormLayoutNotAssigned` error.

### 2.5 Tenant + role resolution at runtime

- **Tenant:** resolved by the existing `ExtractTenant` middleware and pinned via `RunInTenantTx(ctx, pool, …)` (which `SET search_path TO tenant_<code>, public` on the transaction). All queries on `eamform_*` tables MUST go through that helper. The form builder service layer does **not** see a `tenant_id` value, does not pass it as a query parameter, and does not own any tenant-aware logic. See Section 2.0.
- **Role:** read from `c.Locals("activeRole")`, populated by the `ContextLoader` middleware which reads `eamsessions.ses_active_role` (see [architecture conventions](../architecture/conventions.md)). The JWT carries identity only — the active role is **session state**, not a claim. A user has one active role at a time; switching roles requires a frontend menu reload. The handler reads the active assignment for `(frm_key, role_name)`, then reads `flv_definition` of the layout's published version, then returns it. The frontend never sees any other layout's JSON.

### 2.6 Migration strategy

Three migration files, run via `nova-migrate`:

1. **`backend/migrations/tenant/20260219000001_form_definitions.up.sql`** — creates `eamform_definitions`, `eamform_layouts`, `eamform_layout_versions`, `eamform_role_assignments`, `eamform_audit_log`, indexes (including the partial uniques — none of them carry a tenant column because the tables don't have one), the immutability triggers on `eamform_layout_versions` and `eamform_audit_log`, and `CHECK` constraints on `flv_kind`/`fl_status`/`frm_status`/`fal_action`/`fal_entity_type`.
2. **`backend/migrations/tenant/20260219000002_form_submissions.up.sql`** — creates `eamform_submissions` with `fk_layout_version` and `(form_id, created_at DESC)` index.
3. **`backend/migrations/tenant/20260219000003_seed_form_designer_role.up.sql`** — adds the `form_designer` role to the seeded roles, with `formbuilder.view_draft`, `formbuilder.design` permissions.

### 2.7 Note on the `flv_description` field

`flv_description` is the "commit message" of a published layout version. Three rules:

1. **Who fills it:** the human who pressed "Publish". It is **required** on publish (the form does not submit without a non-empty value); it is **optional** on `version.draft_save` (drafts can have placeholder descriptions).
2. **When:** written in the same transaction as the version row insert. The audit row for `version.publish` copies the description into `fal_metadata.description` for full-text search.
3. **What for:** humans, primarily. Reviewers see it on the history list ("added tax_id section") and in the audit log. It is never displayed to end users.

---

## 3. Layout JSON Schema

The persisted JSON in `eamform_layout_versions.flv_definition` is one document per published version, per layout. The schema below is the contract. **Each layout is a complete, standalone JSON document representing the entire UX for one role audience.** There is no per-field role filtering, no policy document, no derived fields. If admin needs a different field than agent, that's a different layout.

### 3.1 Top-level shape

```json
{
  "schemaVersion": "1.0.0",
  "layoutName": "admin-full",
  "displayName": "Admin (Full)",
  "description": "Full intake form for administrators.",
  "sections": [
    { "id": "sec_identity",  "title": "Identity",  "order": 0, "fields": [ ... ] },
    { "id": "sec_address",   "title": "Address",   "order": 1, "fields": [ ... ] },
    { "id": "sec_consent",   "title": "Consent",   "order": 2, "fields": [ ... ] }
  ],
  "rules": [
    {
      "id": "rule_confirm_email",
      "when":  { "path": "sections.sec_identity.fields.confirm_email" },
      "expect": { "equals": { "path": "sections.sec_identity.fields.email" } },
      "message": "Confirmation email must match the primary email."
    }
  ]
}
```

Differences from the previous model: no `key`/`name` (the form's `key` lives on `eamform_definitions`, the layout's `name` lives on `eamform_layouts.fl_name`). `layoutName` here is a denormalized convenience for self-description, not a uniqueness key.

### 3.2 Field shape

Every field inside `sections[].fields[]` has the same envelope:

```json
{
  "id": "legal_name",
  "key": "legal_name",
  "label": "Legal name",
  "type": "text",
  "order": 0,
  "helpText": "As it appears on official documents.",
  "defaultValue": null,
  "placeholder": null,
  "validators": [
    { "kind": "required" },
    { "kind": "minLength", "value": 2 },
    { "kind": "maxLength", "value": 120 }
  ],
  "options": null,
  "ui": { "width": "full", "readOnly": false }
}
```

`type` is one of the catalog in 3.4. `options` is populated only for `select`, `multiselect`, and `radio`. `ui.width` accepts `"full" | "half" | "third"` for grid layout. `ui.readOnly` is a designer-authored attribute (it is part of the layout, not derived from policy); use it when a layout intentionally shows a non-editable field (e.g., the agent-compact layout shows `tax_id` as readOnly).

### 3.3 Example: "customer-intake" / `admin-full` layout (3 sections, 9 fields, with cross-field rule)

This is the richest of the three layouts for the `customer-intake` form. The `agent-compact` and `viewer-readonly` layouts are stored as separate complete JSON documents with their own sections and fields.

```json
{
  "schemaVersion": "1.0.0",
  "layoutName": "admin-full",
  "displayName": "Admin (Full)",
  "description": "Full intake form for administrators. Includes internal notes.",
  "sections": [
    {
      "id": "sec_identity",
      "title": "Identity",
      "order": 0,
      "fields": [
        {
          "id": "legal_name",
          "key": "legal_name",
          "label": "Legal name",
          "type": "text",
          "order": 0,
          "helpText": "As it appears on official documents.",
          "validators": [
            { "kind": "required" },
            { "kind": "minLength", "value": 2 },
            { "kind": "maxLength", "value": 120 }
          ],
          "ui": { "width": "full", "readOnly": false }
        },
        {
          "id": "tax_id",
          "key": "tax_id",
          "label": "Tax ID",
          "type": "text",
          "order": 1,
          "validators": [{ "kind": "required" }, { "kind": "pattern", "value": "^[A-Z0-9]{8,15}$" }],
          "ui": { "width": "half", "readOnly": false }
        },
        {
          "id": "internal_notes",
          "key": "internal_notes",
          "label": "Internal notes",
          "type": "textarea",
          "order": 2,
          "helpText": "Visible to admins and agents only.",
          "validators": [{ "kind": "maxLength", "value": 1000 }],
          "ui": { "width": "full", "readOnly": false }
        },
        {
          "id": "birth_date",
          "key": "birth_date",
          "label": "Date of birth",
          "type": "date",
          "order": 3,
          "validators": [{ "kind": "required" }, { "kind": "maxDate", "value": "today" }],
          "ui": { "width": "half", "readOnly": false }
        },
        {
          "id": "email",
          "key": "email",
          "label": "Email",
          "type": "text",
          "order": 4,
          "validators": [{ "kind": "required" }, { "kind": "email" }],
          "ui": { "width": "half", "readOnly": false }
        },
        {
          "id": "confirm_email",
          "key": "confirm_email",
          "label": "Confirm email",
          "type": "text",
          "order": 5,
          "validators": [{ "kind": "required" }],
          "ui": { "width": "half", "readOnly": false }
        }
      ]
    },
    {
      "id": "sec_address",
      "title": "Address",
      "order": 1,
      "fields": [
        {
          "id": "country",
          "key": "country",
          "label": "Country",
          "type": "select",
          "order": 0,
          "validators": [{ "kind": "required" }],
          "options": {
            "source": "static",
            "choices": [
              { "value": "AR", "label": "Argentina" },
              { "value": "BR", "label": "Brazil" },
              { "value": "UY", "label": "Uruguay" }
            ]
          },
          "ui": { "width": "half", "readOnly": false }
        },
        {
          "id": "city",
          "key": "city",
          "label": "City",
          "type": "text",
          "order": 1,
          "validators": [{ "kind": "required" }],
          "ui": { "width": "half", "readOnly": false }
        },
        {
          "id": "notes",
          "key": "notes",
          "label": "Notes",
          "type": "textarea",
          "order": 2,
          "validators": [{ "kind": "maxLength", "value": 500 }],
          "ui": { "width": "full", "readOnly": false }
        }
      ]
    },
    {
      "id": "sec_consent",
      "title": "Consent",
      "order": 2,
      "fields": [
        {
          "id": "accepts_terms",
          "key": "accepts_terms",
          "label": "I accept the terms of service",
          "type": "checkbox",
          "order": 0,
          "validators": [{ "kind": "requiredTrue" }],
          "ui": { "width": "full", "readOnly": false }
        }
      ]
    }
  ],
  "rules": [
    {
      "id": "rule_confirm_email",
      "when":  { "path": "sections.sec_identity.fields.confirm_email" },
      "expect": { "equals": { "path": "sections.sec_identity.fields.email" } },
      "message": "Confirmation email must match the primary email."
    }
  ]
}
```

### 3.4 Field type catalog (MVP)

Each entry justifies why it ships in v1.

| Type | Purpose | Justification |
|---|---|---|
| `text` | Single-line free text. | Universal. Required for almost every form. |
| `textarea` | Multi-line free text. | Same as above for longer answers. |
| `number` | Numeric input with min/max/integer support. | Common in inventory, finance, configuration. |
| `date` | Calendar picker. | Required for birth dates, deadlines, scheduling. |
| `checkbox` | Boolean (single). | Required for consent, opt-in, flags. |
| `select` | Single-choice dropdown. | Replaces hundreds of radio buttons and saves vertical space. |
| `radio` | Single-choice inline (≤5 options). | When users need to *see* all options to choose. |
| `multiselect` | Multi-choice tag picker. | Tags, categories, permissions. |

**Deliberately out of v1**: `file`/`image`, `richtext`, `repeater`/`matrix`, `lookup-from-API`, `signature`, `computed`. They are real needs but each is a feature on its own and they were not in the requirements.

### 3.5 Cross-field validators

Declared in `top-level.rules[]`. Each rule is a predicate over a path tree; if the predicate fails, the field referenced by `when.path` receives the error message in `message`.

Shape:

```json
{
  "id": "rule_confirm_email",
  "when":  { "path": "sections.sec_identity.fields.confirm_email" },
  "expect": { "equals": { "path": "sections.sec_identity.fields.email" } },
  "message": "Confirmation email must match the primary email."
}
```

Supported operators in v1:

| Operator | Semantics |
|---|---|
| `equals` | Strict equality of values. |
| `notEquals` | Inverse of `equals`. |
| `requiredIf` | Field becomes required when the path has a truthy value. |
| `hiddenIf` | Field becomes hidden when the path has a truthy value. **Server-side applies this.** |

The path syntax is dot-delimited through `sections.<sectionId>.fields.<fieldKey>`.

### 3.6 Sections and ordering

- `sections[].order` is an `int`. Sections are sorted ascending by `order` at resolve time (defensive — the server does not trust array order).
- Within a section, `fields[].order` is the field order. Same defensive sort.
- The designer UI also enforces uniqueness of `order` within a parent (see Section 5.1).
- Reordering between sections is modeled by changing a field's `section_id` and `order` together; the JSON doesn't carry a `section_id` per field because fields live nested under `sections[].fields[]`.

---

## 4. Runtime Resolution Algorithm

When a client calls `GET /api/formbuilder/forms/:formKey`, the backend runs:

1. **Auth + context.** Read `user_id` from JWT claims (`middleware.GetUserClaims(c)`). Read `role_name` from `c.Locals("activeRole")` — this was loaded by the `ContextLoader` middleware from `eamsessions.ses_active_role` (see [architecture conventions](../architecture/conventions.md)). The JWT carries identity only; the active role is session state. Tenant is **not** extracted here — it was already resolved by the `ExtractTenant` middleware and pinned via `search_path` before this handler runs (Section 2.0). The handler never sees a `tenant_id` value.
2. **Permission check.** Call `roles.HasPermission("formbuilder", action)`:
   - `view` — required to read any published layout.
   - `view_draft` — required to read a draft layout.
   - `design` — required for designer endpoints.
   - `publish` — required for publish.
   - `assign` — required for assign layouts to roles.
3. **Look up active role assignment.** Query `eamform_role_assignments` for the active row matching `(fra_form_id, fra_role_name) WHERE fra_revoked_at IS NULL`. Tenant is implicit — the query runs inside `RunInTenantTx`, which has already scoped `search_path` to this tenant's schema.
   - If an active assignment exists → follow steps 4–9 against `fra_layout_id`.
   - If **no** active assignment exists → fall back to the layout named `default` for this form: `SELECT fl_id FROM eamform_layouts WHERE fl_form_id = $1 AND fl_name = 'default'`. If that lookup returns no row (which would be a data-integrity bug, since every form auto-ships a `default` layout), surface `500 FormDefaultLayoutMissing` — this should be impossible in practice but the explicit error helps diagnose a corrupt install. Use the resulting `fl_id` for steps 4–9. **The runtime never returns `FormLayoutNotAssigned`; missing assignment always resolves to `default`.**
4. **Fetch layout.** Load `eamform_layouts` by `fra_layout_id`. If `fl_status = 'archived'` → `410 FormLayoutArchived`.
5. **Fetch published version.** Read `fl_published_version_id` and load that row from `eamform_layout_versions`. If `NULL` → `404 FormLayoutNotPublished`.
6. **Apply cross-field `hiddenIf`.** Walk `flv_definition.rules[]`. For any `hiddenIf` whose condition is true against a `payload` (only relevant for *submission* preview), set `hidden = true` on the target field. For pure form-filling reads (no payload), no rule evaluation happens — the layout is returned as-is.
7. **Re-validate** the document against the Zod schema (`layoutDefinitionSchema`). Safety net — should never fail if the designer validated on save.
8. **Cache.** Write to `formbuilder:layout:{role_name}:{version_id}` with TTL 10 minutes. (The cache key is process-local and lives in a `sync.Map` — see Section 4.3 — so cross-tenant leakage is structurally impossible: each API process serves one tenant at a time per request, and entries are never shared across processes.) Keyed by version so publishing invalidates by version-id change; old entries expire on TTL.
9. **Respond.** Return the layout JSON to the client. The client renders it; it never re-applies role logic.

That is the entire algorithm. Three joins (assignment lookup, `default` fallback lookup on cache miss, version fetch), one cache write, one Zod validation. No policy walk, no field filtering, no `effectiveFields` derivation.

### 4.1 Fallback example: `customer-intake` with three roles, only two assignments

Concrete trace, picking up the schema from Section 2.4. The form `customer-intake` has three layouts: `default` (`fl_id = 100`, published `flv_id = 1009`), `admin-full` (`101` / `1003`), and `agent-compact` (`102` / `1005`). The assignment table holds only two active rows: `admin → 101`, `agent → 102`. There is no row for `viewer`, `manager`, or `auditor`.

| Caller's role | Step 3 (assignment lookup) | Step 3 fallback (since no row) | Steps 4–6 (layout + version) | Returned `flv_id` |
|---|---|---|---|---|
| `admin` | `fra_layout_id = 101` | — | `fl_id = 101` → `fl_published_version_id = 1003` | `1003` |
| `agent` | `fra_layout_id = 102` | — | `fl_id = 102` → `fl_published_version_id = 1005` | `1005` |
| `viewer` | no row | `default` lookup → `fl_id = 100` | `fl_id = 100` → `fl_published_version_id = 1009` | `1009` |
| `manager` | no row | `default` lookup → `fl_id = 100` | `fl_id = 100` → `fl_published_version_id = 1009` | `1009` |
| `auditor` | no row | `default` lookup → `fl_id = 100` | `fl_id = 100` → `fl_published_version_id = 1009` | `1009` |

Three roles — `viewer`, `manager`, `auditor` — get the same `default` layout (v2, `flv_id = 1009`), even though none of them has a row in `eamform_role_assignments`. The resolver never tells the caller "no assignment"; it just returns `default`. The cache key is `formbuilder:layout:{viewer|manager|auditor}:1009` — three distinct keys for the same payload. That is intentional: invalidating one role's view (TTL expiry, publish) should not invalidate the others, even though the underlying JSON is identical. The cache lives inside the API process that already serves only this tenant (Section 2.0), so no tenant prefix is needed.

**Audit trail for fallback resolutions.** The runtime resolution path is a read, not a write — it does not write to `eamform_audit_log`. Designers can audit "who saw what when" via `eamform_submissions.fs_layout_id` once submissions are enabled (Section 7 PR6). For a read-only fallback trace, future work could log a separate `form.resolve` event into `eamform_audit_log` with `entity_type = 'form'` and `metadata = { role_name, layout_id, fallback_used: true|false }`; this is **not** in v1 scope.

### 4.2 Server-side vs client-side split

| Concern | Server | Client |
|---|---|---|
| Resolve layout via assignment + version | ✅ | — |
| Render HTML | — | ✅ |
| Field-level validator execution (length, pattern, min/max) | — | ✅ (Zod, same schema) |
| Cross-field `equals` / `requiredIf` | — | ✅ (evaluated against FormGroup values) |
| Cross-field `hiddenIf` (preview/submit only) | ✅ | — |
| Submission to backend | — | ✅ |

The asymmetry is intentional and narrower than before: **layout selection is server-authoritative; UX validation is client-immediate.** The backend does not re-run cross-field rules on submission in v1 — field-level validators are sufficient for the user's requirements.

### 4.3 Caching strategy

- **Key:** `formbuilder:layout:{role_name}:{version_id}`. The key includes the **caller's role name**, not the resolved layout name — so fallback resolutions for `viewer`/`manager`/`auditor` land on three distinct cache entries even when they all return the same `default` version (see Section 4.1). This is fine: the duplicate storage is a few hundred bytes per role and avoids accidental cross-role invalidation when one role is later assigned a different layout.
- **Storage:** in-process `sync.Map` cache in the API process, wrapped in a `LayoutCache` interface so a Redis backend can be dropped in later.
- **TTL:** 10 minutes. Publishing produces a new `version_id`, which produces a new cache key; old keys expire on TTL. We do not invalidate by pattern.
- **Per-tenant isolation:** the API process serves one tenant per request via `search_path`, and the in-process `sync.Map` is not shared across processes. Cache entries cannot leak across tenants by construction — no `tenant_id` prefix is required. If a future Redis backend is added, the key would need a tenant segment at that point (or, better, one Redis instance per tenant).
- **Drafts:** not cached (drafts change frequently and are accessed by few users).
- **Assignments:** not cached as a separate entity — the assignment lookup is a single indexed query, cheap enough to skip caching.
- **`default` fallback lookups:** cached as part of the assignment-lookup flow (the resolved layout is cached; the fact that it came from `default` is not). On a publish of `default`, the cache key bumps to a new `version_id` and stale entries expire on TTL — same invalidation story as any other layout.

---

## 5. Frontend Architecture

### 5.1 Component tree

All components are **standalone**, signal-based, and live under `frontend/src/app/features/form-builder/`.

```
features/form-builder/
├── designer/
│   ├── form-designer.component.ts        # The visual builder UI.
│   ├── form-designer.component.html
│   ├── form-designer.component.css
│   ├── layout-picker.component.ts        # NEW: lists existing layouts for the form;
│   │                                     # lets the user pick one to edit or create new.
│   ├── field-palette.component.ts        # Sidebar with field types to drag in.
│   ├── section-canvas.component.ts       # The middle canvas; hosts sections + drop zones.
│   ├── section-card.component.ts         # One section, hosts draggable fields.
│   ├── field-card.component.ts           # One field row.
│   ├── field-settings.component.ts       # Right-side panel: validators, label, helpText.
│   ├── preview-dialog.component.ts       # Opens FormRuntimeComponent in preview mode.
│   └── assignment-panel.component.ts     # NEW: lists roles and the layout each is currently
│                                         # assigned to; lets the designer change the mapping.
├── runtime/
│   ├── form-runtime.component.ts         # The renderer used by end users.
│   ├── form-runtime.component.html
│   ├── form-runtime.component.css
│   ├── form-section.component.ts         # Renders one section (fieldset-style).
│   ├── field-renderers/
│   │   ├── field-renderer.component.ts   # Switch over `type`, delegates to a concrete renderer.
│   │   ├── text-field.component.ts
│   │   ├── textarea-field.component.ts
│   │   ├── number-field.component.ts
│   │   ├── date-field.component.ts
│   │   ├── checkbox-field.component.ts
│   │   ├── select-field.component.ts
│   │   ├── radio-field.component.ts
│   │   └── multiselect-field.component.ts
│   └── cross-field-validator.directive.ts # Applies `rules[]` against FormGroup changes.
├── state/
│   ├── form-designer.store.ts            # Signal-based local state.
│   └── form-runtime.store.ts             # Signal-based local state (values, errors).
├── models/
│   ├── layout-definition.model.ts        # TS types mirroring Section 3.
│   ├── assignment.model.ts               # Role-to-layout mapping shape.
│   └── audit-entry.model.ts              # Audit log entry shape.
├── services/
│   ├── form-designer.service.ts          # HTTP client for the designer.
│   ├── form-runtime.service.ts           # HTTP client for the renderer.
│   └── assignment.service.ts             # HTTP client for the assignment UI.
└── form-builder.routes.ts                # Lazy-loaded route config.
```

#### `FormDesignerComponent`

- Loads the form's layouts via `GET /api/formbuilder/forms/:formKey/layouts`. The `LayoutPickerComponent` (new) lists them with their `name`, `displayName`, status, and draft/published pointers, and lets the user pick one to edit or create a new one. The layout named `default` appears with a visible **"Default" badge** and the helper text "Applies to all roles without a specific assignment" — designers see at a glance which one is the system-managed fallback. The create flow asks for a unique slug-friendly `name` and rejects duplicates via the backend's unique constraint; it also rejects `default` as a user-chosen name (the backend returns `409 ReservedLayoutName`) since the system already created one when the form was inserted.
- Once a layout is picked, loads its draft (or creates one on first edit) via `GET /api/formbuilder/forms/:formKey/layouts/:layoutName/draft`.
- Renders three columns: `FieldPaletteComponent` (left) · `SectionCanvasComponent` (middle) · `FieldSettingsComponent` (right, conditional on selection).
- Drag & drop: **Angular CDK DragDrop** (`@angular/cdk/drag-drop`) — already used by `shared/components/query-builder`, so we follow the established convention. We use a single `cdkDropList` per section plus a global one for the palette; `transferArrayItem` handles cross-section moves; `moveItemInArray` handles reorders.
- State lives in `FormDesignerStore` (signals): `currentLayout`, `sections`, `selectedFieldId`, `dirty`. Saving serializes the store to JSON, validates with Zod, and `PUT`s to the backend.
- Save flow:
  1. `validate(layoutDefinitionSchema)` locally — show inline errors if any.
  2. `PUT /api/formbuilder/forms/:formKey/layouts/:layoutName/draft` with the JSON.
  3. Backend validates again, persists a new draft row if the pointer is `NULL`, updates the pointer otherwise.
  4. On 200, store updates `lastSavedAt`. No auto-save in v1 — explicit "Save draft" and "Publish" buttons.
- Publish flow:
  1. `POST /api/formbuilder/forms/:formKey/layouts/:layoutName/publish` with `description` (commit message).
  2. Backend inserts a new `published` version row, moves the `fl_published_version_id` pointer.
  3. Designer UI shows the new version number and refreshes the history panel.
- Preview flow: opens `PreviewDialogComponent` which mounts `FormRuntimeComponent` with the current draft in-memory (no backend round-trip) — uses the same renderer code path, which is the point.
- Assignment UI (`AssignmentPanelComponent`, new): lists each role and the layout currently assigned to it; lets the admin pick a layout from the form's existing layouts (cannot assign a layout from another form). Roles that have **no** active assignment are shown with the placeholder "→ uses `default` layout" so designers see the implicit fallback. Save flow: `PUT /forms/:formKey/assignments/:roleName` with `{ layoutName }`. Designers can also explicitly assign the `default` layout to a role — this is redundant but legal, and the panel shows it as "default (intentional)" so the assignment table stays self-documenting; alternatively, designers can simply leave the row empty and rely on the fallback rule. **Recommendation: leave it empty** — fewer rows in the assignments table, and the panel's placeholder already conveys the intent.

#### `FormRuntimeComponent`

- Consumes a **layout JSON** document (returned by `GET /api/formbuilder/forms/:formKey` — the backend has already resolved which layout to serve for the caller's role).
- On `init`, calls `FormBuilder.group()` (Angular Reactive Forms) to construct a `FormGroup` keyed by `field.key`, with `FormControl`s carrying validators translated from `field.validators[]` (via the same Zod schema used on the backend).
- Iterates `sections[]` (sorted by `order`) and renders each via `FormSectionComponent`.
- For each field, delegates to `FieldRendererComponent` which switches on `field.type`. `ui.readOnly` (a designer-authored attribute on the field, not policy) is honored at render time.
- A `valueChanges` subscription on the `FormGroup` runs `cross-field-validator.directive` to evaluate `rules[]` and set errors on the right controls.
- On submit, validates the entire `FormGroup`; if valid, `POST`s `{ versionId, payload }` to `/api/formbuilder/forms/:formKey/submissions`. Errors are displayed inline per field and as a toast summary.

#### `FieldRendererComponent`

Conventions:

- One concrete component per field type, all implementing a common `FieldRender` interface: `value: Signal<unknown>`, `setValue(v)`, `error: Signal<string | null>`.
- The switch in `FieldRendererComponent` is a `computed()` over the field's `type` that returns the right component reference. New field types ship as a new file + a new entry in the switch.
- All field components consume `ControlValueAccessor` semantics via Angular Reactive Forms directly — no custom CVA needed.

#### `FormSectionComponent`

Renders a `<fieldset>` with the section title and a list of fields. Toggling `hidden` on the section itself (a future capability) drops it from the DOM entirely.

### 5.2 State management

Signal-based local state in `FormDesignerStore` and `FormRuntimeStore`. **No NgRx.** Rationale:

- The form builder is feature-scoped; cross-feature state sharing is minimal (forms are not embedded into other screens in v1).
- Signals + `computed()` are sufficient for the reactive flow inside the designer.
- The Nova codebase already uses signals everywhere — adding NgRx would be a one-off divergence.

If submissions are later embedded into a dashboard or search results, we revisit with `SignalStore` (`@ngrx/signals`) before reaching for full NgRx.

### 5.3 Folder structure

```
frontend/src/app/features/form-builder/
├── designer/        (see 5.1)
├── runtime/         (see 5.1)
├── state/           (signal stores)
├── models/          (TS types)
├── services/        (HTTP clients)
└── form-builder.routes.ts
```

Registered in `frontend/src/app/app.routes.ts` under a lazy `loadChildren: () => import('./features/form-builder/form-builder.routes')`.

### 5.4 Integration with existing patterns

- **TanStack tables** are unaffected — forms are not rendered inside grid rows in v1.
- **Zod**: the same schemas from `shared/form-schemas/` (see Section 6.4) are imported directly by both the designer (for save-time validation) and the runtime (for control-level validator generation).
- **Multi-tenant**: the existing auth context provider already routes the request to the correct tenant's schema (Section 2.0). `FormRuntimeService` makes the same HTTP calls it would for any other module — no `tenant_id` is added to URLs, headers, or bodies. If a future request-routing change moves to a "tenant selector in the designer UI" model, that is a cross-cutting change owned by the platform team, not this module. **There is no tenant selector in the designer in v1**; designers always operate against the tenant implied by their session.
- **`query-builder` clarification**: `frontend/src/app/shared/components/query-builder/` is the existing visual builder for grid **queries** (filters/sorts/columns) — it uses the same Angular CDK DragDrop and signals patterns we will follow, but solves a different problem. It is **not** a form builder and there is no overlap to consolidate.

---

## 6. Backend Architecture

### 6.1 Package boundaries

```
backend/internal/
├── domain/
│   └── formbuilder/
│       ├── entity.go              # Form, Layout, LayoutVersion, Assignment, AuditEntry structs.
│       ├── ports.go               # FormRepository, LayoutRepository, LayoutVersionRepository,
│       │                          # AssignmentRepository, AuditLogRepository, FormPermissionChecker.
│       ├── service.go             # Business logic: create form, create layout, save draft,
│       │                          # publish, assign, archive, resolve.
│       ├── dto.go                 # HTTP DTOs (request/response shapes).
│       ├── errors.go              # Domain error sentinels.
│       └── resolve.go             # The resolution algorithm from Section 4.
├── adapters/
│   ├── api/
│   │   └── formbuilder/
│   │       ├── handler.go         # Fiber handlers.
│   │       ├── routes.go          # Route registration.
│   │       └── middleware.go      # Form-builder-specific auth gate.
│   └── db/
│       └── formbuilder/
│           ├── form_repository.go
│           ├── layout_repository.go
│           ├── layout_version_repository.go
│           ├── assignment_repository.go
│           └── audit_log_repository.go
└── infrastructure/
    └── cache/
        └── layout_cache.go         # In-process cache implementation (replaces resolved_form_cache).
```

Hexagonal boundaries:

- `domain/formbuilder` has **zero** imports from `adapters/` or `infrastructure/`.
- `adapters/db/formbuilder` implements the `domain/formbuilder.*Repository` interfaces; uses `infrastructure/db.RunInTenantTx` for every transaction (so the tenant's `search_path` is already set, and SQL queries never reference a `tenant_id` column).
- `adapters/api/formbuilder` mounts the Fiber routes; calls into `domain/formbuilder.Service`. The Fiber handler does **not** read a `tenant_id` from the request and does **not** pass it down — tenant is implicit in the database connection that `RunInTenantTx` opens.
- The audit log writes happen inside the service layer (`CreateLayout` writes both `eamform_layouts` and `eamform_audit_log` in one transaction).

### 6.2 API endpoints

All endpoints under `/api/formbuilder`. All require authentication; role-gated actions noted.

| Method | Path | Purpose | Auth |
|---|---|---|---|
| `GET`    | `/api/formbuilder/forms` | List form definitions in the current tenant. | `formbuilder.view` |
| `POST`   | `/api/formbuilder/forms` | Create a new form definition. Body: `{ key, name, description }`. **Side effect:** in the same transaction, auto-creates a layout named `default` (`fl_status = 'active'`, `fl_draft_version_id = NULL`, `fl_published_version_id = NULL`) and writes a `layout.create` audit entry with `metadata.layout_name = "default"` and `metadata.auto_created = true`. | `formbuilder.design` |
| `POST`   | `/api/formbuilder/forms/:formKey/archive` | Soft-archive a form. | `formbuilder.publish` |
| `GET`    | `/api/formbuilder/forms/:formKey/layouts` | List layouts for a form (name, status, draft/published pointers). | `formbuilder.view` |
| `POST`   | `/api/formbuilder/forms/:formKey/layouts` | Create a new layout. Body: `{ name, displayName, description }`. | `formbuilder.design` |
| `POST`   | `/api/formbuilder/forms/:formKey/layouts/:layoutName/archive` | Soft-archive a layout. **Refuses** to archive the layout named `default` while the parent form is `active`; the form must be archived first. Once the form is archived, the `default` layout may also be archived (but never hard-deleted — submissions are still bound to it). | `formbuilder.publish` |
| `GET`    | `/api/formbuilder/forms/:formKey/layouts/:layoutName/draft` | Get the **draft** (designer-only). | `formbuilder.view_draft` |
| `PUT`    | `/api/formbuilder/forms/:formKey/layouts/:layoutName/draft` | Save the current draft. Body: the full layout JSON. | `formbuilder.design` |
| `POST`   | `/api/formbuilder/forms/:formKey/layouts/:layoutName/publish` | Promote the current draft to a new published version. Body: `{ description }`. | `formbuilder.publish` |
| `POST`   | `/api/formbuilder/forms/:formKey/layouts/:layoutName/revert` | Set the draft to be a copy of a past published version. Body: `{ versionNumber }`. | `formbuilder.design` |
| `GET`    | `/api/formbuilder/forms/:formKey/layouts/:layoutName/versions` | Version history. | `formbuilder.view` |
| `GET`    | `/api/formbuilder/forms/:formKey/layouts/:layoutName/versions/:n` | Get a specific historical version. | `formbuilder.view` |
| `GET`    | `/api/formbuilder/forms/:formKey/assignments` | List role-to-layout assignments for the form. | `formbuilder.view` |
| `PUT`    | `/api/formbuilder/forms/:formKey/assignments/:roleName` | Assign a layout to a role. Body: `{ layoutName }`. Replaces any previous active assignment for `(form, role)`. | `formbuilder.assign` |
| `DELETE` | `/api/formbuilder/forms/:formKey/assignments/:roleName` | Revoke an assignment (soft-delete, sets `fra_revoked_at`). | `formbuilder.assign` |
| `GET`    | `/api/formbuilder/forms/:formKey/audit` | List audit entries for a form (paged, filterable by action/entity). | `formbuilder.view` |
| `GET`    | `/api/formbuilder/forms/:formKey` | **Public runtime resolution.** Resolves the form for the caller's role and returns the layout JSON. | `formbuilder.view` |
| `POST`   | `/api/formbuilder/forms/:formKey/submissions` | Submit a filled form. Body: `{ versionId, payload }`. | `formbuilder.view` |

### 6.3 Authorization model

Uses the generic `roles.HasPermission(screen, action)` pattern (see [architecture conventions](../architecture/conventions.md) §5). Permissions use **semantic actions**, not the legacy CRUD column model. The target schema is `eamrole_permissions(role, screen, action, allowed)` — a normalized table where each row grants (or denies) a specific action on a specific screen. The `Role.HasPermission` method already supports this, including wildcards (`*` screen or `*` action).

The active role is loaded from `eamsessions.ses_active_role` by the `ContextLoader` middleware and placed in `c.Locals("activeRole")`. The handler loads the role entity from DB, then checks:

```go
activeRole := c.Locals("activeRole").(string)
role := loadRole(activeRole)
if !role.HasPermission("formbuilder", "design") {
    return c.Status(403).JSON(...)
}
```

**Permission matrix:**

| Action | Description | Granted to |
|---|---|---|
| `formbuilder.view` | Read published layouts for filling; view version history and audit log. | All authenticated users (most roles). |
| `formbuilder.view_draft` | Read drafts. | `admin`, `form_designer`. |
| `formbuilder.design` | Create forms, create layouts, edit drafts. | `admin`, `form_designer`. |
| `formbuilder.publish` | Promote draft → published; archive forms or layouts. | `admin` only. |
| `formbuilder.assign` | Bind a layout to a role for a form. | `admin` only. |

A `form_designer` role needs to be added to the seeded roles (see migration `20260219000003_seed_form_designer_role.up.sql`). Permissions are inserted into `eamrole_permissions` with semantic actions (e.g. `(form_designer, formbuilder, design, true)`, `(form_designer, formbuilder, view_draft, true)`), not the legacy CRUD columns.

### 6.4 Service-layer invariants for the `default` layout

The `default` reserved name is enforced entirely in `domain/formbuilder/service.go`, not in SQL. The service is the only writer of `eamform_layouts` rows and is the place where the cross-table invariants live. Concretely:

- **`CreateForm(tx, …)`** writes the `eamform_definitions` row, then in the same transaction writes the `eamform_layouts` row with `fl_name = "default"`, `fl_status = "active"`, `fl_draft_version_id = NULL`, `fl_published_version_id = NULL`. No draft version is created at this point — the layout exists as a placeholder and the designer can publish an empty `default` later if they want. The corresponding `layout.create` audit row carries `metadata.auto_created = true`.
- **`CreateLayout(tx, formID, name, …)`** rejects `name == "default"` with `409 ReservedLayoutName` and a message pointing to the auto-created layout. The error is the same `errors.go` sentinel used everywhere else.
- **`ArchiveLayout(tx, formID, name)`** rejects `name == "default"` with `409 CannotArchiveDefault` while the parent form is `active`. Once the form is archived, the call succeeds.
- **`HardDeleteForm(tx, formID)`** (admin-only path, not exposed via API in v1 — it exists for compliance tools) is gated by a DB trigger (see Section 2.3) but the service layer also asserts "every form has exactly one `default` layout" in the pre-delete read so the trigger is a belt-and-braces, not the only line of defense.
- **`Resolve(ctx, formKey, roleName)`** runs the algorithm in Section 4. **Note: no `tenantID` parameter** — the tenant is already pinned by the `RunInTenantTx` middleware that opened the transaction the service is operating on. The fallback branch performs the secondary lookup of the `default` layout; if that lookup returns no row, the function returns `500 FormDefaultLayoutMissing` (a sentinel in `errors.go`). The error is logged at `error` level with `form_key` and `role_name` so a missing `default` is loud in observability even though it is unreachable in practice. (For cross-cutting observability — metrics, traces — the request's `tenant` is available in the logger's context, but it is not passed as a service-layer argument.)

### 6.5 Validation strategy: shared Zod schemas

**Canonical source location:** a new top-level directory `shared/form-schemas/` at the repo root. It will be imported by:

- The Go backend (via `go:embed` — the JSON Schema is embedded into the binary at compile time, then loaded into a Zod-equivalent Go validator; this is the canonical approach for Go+Zod pairing and avoids drift because the JSON file is the artifact, not the Go struct).
- The Angular frontend (via `tsconfig` path mapping `"@shared/form-schemas": ["shared/form-schemas/index.ts"]`).

Files in `shared/form-schemas/`:

```
shared/form-schemas/
├── package.json            # name: "@nova/form-schemas", exports the TS types.
├── layout-definition.schema.ts   # Section 3 Zod schema (one per layout, complete JSON).
├── assignment.schema.ts          # Role-to-layout assignment shape.
├── audit-entry.schema.ts         # Audit log entry shape.
├── field-types.ts                # Discriminated union for `type`.
├── validators.ts                 # Validator kind catalog (required, minLength, …).
├── rules.ts                      # Rule/operator catalog.
├── index.ts                      # Public re-exports.
└── schema-version.ts             # Bumped on breaking changes.
```

**Why a TS-first schema with `go:embed`**, not native Go structs: the user explicitly named Zod as the shared validation library. Using TS as the source keeps both sides literally identical, and `go:embed` lets us read the file at compile time and validate with a hand-written Go validator (or generate one later). The alternative — duplicating the schema in Go — is rejected because drift is the failure mode we are explicitly trying to avoid.

The frontend already imports Zod-derived types via `frontend/src/mocks/queries.ts`; we extend that pattern by promoting shared schemas to their own package.

---

## 7. Implementation Plan (PR breakdown)

Six PRs in dependency order. **Total estimated diff: ~3,350 lines** (down from ~3,450 in the previous revision: removing `tenant_id` columns and per-tenant unique indexes trims ~100 LOC across the migration, entities, ports, repositories, and cache key formatters). This still significantly exceeds the 400-line threshold — **PR chaining is recommended.** The `default`-layout fallback rule is folded into PR1 (it touches the schema, the service-layer invariants, and the public resolve endpoint — splitting it would force a second migration on a freshly created table, which is worse than a slightly larger first PR).

The split differs from the previous plan because the layout model now has three independent axes (forms, layouts, assignments) that need to ship in their own reviewable units:

- **Chain 1: Backend foundation (PR1).** Backend-only, ships the DB schema + domain layer + read endpoint. ~600 LOC.
- **Chain 2: Backend design + publish (PR2).** Backend-only, builds the create/publish/assign endpoints on top of Chain 1. ~550 LOC.
- **Chain 3: Backend audit log (PR3).** Backend-only, ships the audit endpoint and confirms all service-layer writes go through the audit helper. ~200 LOC.
- **Chain 4: Frontend foundation (PR4).** Frontend-only, ships the runtime renderer reading the resolved document. ~700 LOC.
- **Chain 5: Frontend designer (PR5).** Frontend-only, ships the visual builder UI including the layout-selector and assignment UI. ~950 LOC.
- **Chain 6: Integration + cross-field rules + polish (PR6).** Frontend + backend, wires shared schemas end-to-end and ships cross-field validators. ~400 LOC.

This ordering keeps each chain reviewable in isolation and keeps backend merged before frontend depends on it. Audit is its own PR because its surface is small but its correctness is critical (the immutability trigger is non-trivial and deserves focused review).

> **PR1 saves ~50 LOC compared to the previous revision** because the form builder module no longer adds any tenant plumbing: no `tenant_id` columns in entities, no `tenant_id` parameter in repository methods, no `tenant_id` predicate in SQL, no `tenant_id` segment in cache keys. The `RunInTenantTx` helper already isolates the schema — repositories just inherit the scoped transaction.

### PR1 — Backend foundation: schema, domain, read endpoint

- **Scope:** backend only.
- **Ships:**
  - Migration `20260219000001_form_definitions.up.sql` (creates `eamform_definitions`, `eamform_layouts`, `eamform_layout_versions`, `eamform_role_assignments`, `eamform_audit_log` — **no `tenant_id` columns**; indexes are scoped per-schema, partial uniques (e.g. `(fl_form_id, fl_name)`, `(flv_layout_id) WHERE flv_kind = 'draft'`, `(fra_form_id, fra_role_name) WHERE fra_revoked_at IS NULL`), immutability triggers on versions and audit log, `CHECK` constraints, and the `BEFORE DELETE` trigger on `eamform_definitions` that rejects deleting a form which still has a `default` layout).
  - `domain/formbuilder/{entity,ports,errors,resolve,service,dto}.go` — including the `default`-layout service-layer invariants from Section 6.4 (`ReservedLayoutName`, `CannotArchiveDefault`, `FormDefaultLayoutMissing`). **No `tenant_id` field on any entity struct**; the `Resolve` signature is `(ctx, formKey, roleName)` (Section 6.4).
  - `adapters/db/formbuilder/{form_repository,layout_repository,layout_version_repository,assignment_repository,audit_log_repository}.go` — all queries run inside `RunInTenantTx`; SQL uses no `tenant_id` predicate.
  - `adapters/api/formbuilder/{handler,routes,middleware}.go` for `GET /forms`, `POST /forms` (auto-creates the `default` layout in the same transaction), `GET /forms/:formKey` (public resolve — implements the Section 4 fallback to `default`), `GET /forms/:formKey/layouts`, `POST /forms/:formKey/layouts` (rejects `name = "default"`), `GET /forms/:formKey/layouts/:layoutName/draft`, `GET /forms/:formKey/assignments`. **No `tenant_id` is read from the request or passed to the service layer.**
  - `infrastructure/cache/layout_cache.go` — key shape `formbuilder:layout:{role_name}:{version_id}` (no tenant segment).
  - Migration `20260219000003_seed_form_designer_role.up.sql`.
- **LOC delta:** ~600.
- **Verified by:** `go build`, `golangci-lint run`, integration tests: list forms, create form, assert a `default` layout exists automatically, create two additional layouts (`admin-full`, `agent-compact`), assign them to `admin`/`agent`, resolve the form for `admin`/`agent` (assigned) and for `viewer`/`manager`/`auditor` (unassigned — must fall back to `default`), assert each response is the right complete layout JSON. Negative tests: creating a layout named `default` returns `409 ReservedLayoutName`; archiving the `default` layout while the form is `active` returns `409 CannotArchiveDefault`.

### PR2 — Backend design and publish endpoints

- **Scope:** backend only.
- **Ships:**
  - Endpoints: `PUT /forms/:formKey/layouts/:layoutName/draft`, `POST /forms/:formKey/layouts/:layoutName/publish`, `POST /forms/:formKey/layouts/:layoutName/revert`, `POST /forms/:formKey/layouts/:layoutName/archive`, `POST /forms/:formKey/archive`, `GET /forms/:formKey/layouts/:layoutName/versions`, `GET /forms/:formKey/layouts/:layoutName/versions/:n`.
  - `PUT /forms/:formKey/assignments/:roleName`, `DELETE /forms/:formKey/assignments/:roleName`.
  - Audit writes for every mutation (form create, layout create/draft-save/publish/archive/assign/unassign) implemented in the service layer.
- **LOC delta:** ~550.
- **Verified by:** integration tests: create form → create 3 layouts → save drafts → publish each → assign → resolve for each role → check `eamform_audit_log` rows exist with correct actor/action/entity_id/metadata. Then re-publish one layout → assert prior version is still queryable by `versionNumber`.

### PR3 — Backend audit log endpoint

- **Scope:** backend only (small).
- **Ships:**
  - `GET /forms/:formKey/audit` (paged, filterable by `action`, `entity_type`, `actor_user_id`, date range).
  - `AuditLogRepository.List` with the corresponding filters and pagination.
  - Tests for immutability: attempts to `UPDATE` or `DELETE` an audit row raise.
- **LOC delta:** ~200.
- **Verified by:** integration test that drives the full mutation flow from PR2 and asserts each mutation wrote exactly one audit row with correct shape; separate test that confirms the immutability trigger raises on `UPDATE` and `DELETE`.

### PR4 — Frontend foundation: runtime renderer

- **Scope:** frontend only.
- **Ships:**
  - `features/form-builder/runtime/` — `FormRuntimeComponent`, `FormSectionComponent`, `FieldRendererComponent` + 8 concrete field renderers, `CrossFieldValidatorDirective`, `FormRuntimeStore`, `FormRuntimeService`, `models/`, `form-builder.routes.ts` with one lazy route `forms/:formKey` rendering `FormRuntimeComponent`.
  - `shared/form-schemas/` package wired into the Angular `tsconfig`.
  - Reactive Forms + Zod validator generator.
- **LOC delta:** ~700.
- **Verified by:** `pnpm run build`, `pnpm run lint`, `ng test`, MSW-based component test that mounts `FormRuntimeComponent` with a fixture layout JSON (no role filtering on the client — just render what's served) and asserts FormGroup state after typing.

### PR5 — Frontend designer (with layout selector + assignment UI)

- **Scope:** frontend only.
- **Ships:**
  - `features/form-builder/designer/` — `FormDesignerComponent`, `LayoutPickerComponent` (new: lists existing layouts for the form, lets the user pick one to edit or create a new one with a unique slug-friendly name), `FieldPaletteComponent`, `SectionCanvasComponent`, `SectionCardComponent`, `FieldCardComponent`, `FieldSettingsComponent`, `PreviewDialogComponent`, `AssignmentPanelComponent` (new: lists roles and the layout each is currently assigned to; lets the designer change the mapping), `FormDesignerStore`, `FormDesignerService`, `AssignmentService`.
  - Drag & drop wiring with Angular CDK DragDrop.
  - Save-draft and publish flows (calling PR2's endpoints); slug-friendly name input with live uniqueness validation against the backend.
- **LOC delta:** ~950.
- **Verified by:** `pnpm run build`, `pnpm run lint`, `ng test`, MSW-based tests: simulate a drag-reorder and assert the serialized JSON reflects the new order; create a layout named `manager-overview` and assert the backend's uniqueness check rejects a duplicate.

### PR6 — Integration, cross-field rules, end-to-end polish

- **Scope:** frontend + backend (small).
- **Ships:**
  - `rules[]` evaluator on the frontend (cross-field `equals`, `requiredIf`).
  - `hiddenIf` server-side evaluation (only meaningful on submit/preview).
  - `POST /forms/:formKey/submissions` endpoint and `eamform_submissions` migration.
  - Round-trip e2e test: design three layouts, publish each, assign to roles, fill form as agent (asserts `agent-compact` is what was served), submit, verify audit log has all the right entries.
- **LOC delta:** ~400.
- **Verified by:** end-to-end test, manual smoke against a local `nova-migrate`-seeded DB, full `go test ./...` and `ng test`.

---

## 8. Open Questions

These are decisions where the requirements admit more than one defensible answer and we need a human call before locking the spec.

1. **What does the auto-created `default` layout contain at form-creation time, and can `default` itself be unassigned from the system?** When `POST /forms/:formKey` runs, the service creates the `default` layout row with no draft and no published version. The designer's first experience of that layout is an empty canvas. Options:
   - (a) **Empty skeleton** — the row exists, but `fl_draft_version_id` and `fl_published_version_id` are both `NULL` until a designer publishes the first version. Until then, unassigned-role resolutions return `404 FormLayoutNotPublished` (a per-layout "never published" error, NOT the per-role "no assignment" error).
   - (b) **Sensible starter JSON** — the service writes a `draft` containing a single "Hello world" section with one read-only field, and a `published` snapshot of the same. Designers edit the draft instead of starting from zero.
   - (c) **Cloned from a template** — a tenant-level "default-form-template" layout exists, and the service clones it (deep copy of JSON, fresh `fl_id`) into the new form's `default` slot.
   - **Recommendation:** (b). (a) is too sparse and surfaces `FormLayoutNotPublished` for any role that hits the form right after creation — confusing and contradicts the "always serve `default`" promise. (c) is nice for large tenants but introduces a tenant-level template concept we don't otherwise need.
   - **Sub-question:** can `default` itself be unassigned from the system, or is it always present? **Answer:** `default` is always present in every form (the trigger on `eamform_definitions` and the service-layer invariant both enforce this). There is no "delete `default`" path. Designers can leave its assignment table empty (recommended — fallback covers it) or explicitly assign it to a role (legal but redundant).
   - **Tradeoff:** (a) is the strictest interpretation of "always there" but punishes early access; (b) ships a small starter that needs reviewing; (c) is the most polished but the largest scope.

2. **Layouts for roles that don't exist yet.** When a new role is added to a tenant (e.g., a new `auditor` role), no layouts are assigned to it for any form. Should the system (i) auto-create empty placeholder layouts for the new role, (ii) leave them unassigned and rely on the assignment UI, or (iii) clone layouts from a sibling role as a starter?
   - **Recommendation:** (ii). Auto-creation is too magical and clone-on-role-create hides intent. The admin should explicitly assign or create.
   - **Tradeoff:** zero magic vs slower onboarding for tenants with many roles.

3. **Draft across layouts sharing a form.** If a designer opens the draft for `admin-full` while another designer opens the draft for `agent-compact`, both can edit independently. Is that desired, or should the whole form be locked when any layout is being edited?
   - **Recommendation:** independent drafts. Layouts are independent artifacts; cross-layout atomicity is a rare need and complicates the designer's mental model.
   - **Tradeoff:** simpler designer flow vs occasional "I changed admin but forgot agent" reviews.

4. **Submission storage.** We assumed `eamform_submissions` with `jsonb payload` for v1, but the user did not request persistence at all. Options:
   - (a) Ship submissions in v1 (current plan).
   - (b) Defer to v2 — submissions are POSTed to a webhook or just validated and acknowledged.
   - **Recommendation:** defer. Forms are usually a data-capture step; the *next* feature decides what to do with the data. v1 should validate and acknowledge.
   - **Tradeoff:** (a) is needed immediately if any real form must produce records; (b) keeps scope tight.

5. **Permission key naming.** We proposed `formbuilder.view`, `formbuilder.view_draft`, `formbuilder.design`, `formbuilder.publish`, `formbuilder.assign`. Alternatives: per-form permissions (`formbuilder.<form_key>.design`) for finer-grained control.
   - **Recommendation:** ship with the five-action model; add per-form granularity in v2 if a tenant asks.
   - **Tradeoff:** simpler v1 vs future-proofing for delegated designers who should only edit specific forms.

6. **Cache invalidation on publish.** Current plan lets TTLs expire naturally (10 min). Alternatives: a Redis pub/sub channel on publish that flushes the in-process cache across replicas.
   - **Recommendation:** in-process cache for v1, accept up to 10 min of staleness on replicas. Document the limitation.
   - **Tradeoff:** zero infra vs stronger consistency for hot-reload scenarios.

7. **Static vs dynamic `select` options.** Section 3 lists `source: "static"` with inline choices. Dynamic `source: "api"` (e.g., populate from `eamsyscodes`) was deferred.
   - **Recommendation:** ship static-only in v1; mark `options.source` as `"static" | "api"` in the schema so adding `"api"` later is non-breaking.
   - **Tradeoff:** simpler renderer vs re-implementing for every list-type form.

8. **Reorder UX between sections.** Angular CDK DragDrop supports `transferArrayItem` cleanly, but the UX for "drag from section A to section B" without a visible drop placeholder can feel fragile.
   - **Recommendation:** keep CDK DragDrop; require explicit drop zones (sections have visible drop areas between fields, and a dedicated drop zone at the bottom for empty-section inserts).
   - **Tradeoff:** implementation simplicity vs more sophisticated UX (e.g., nested sections, accordion preview).

9. **Localized labels and validator messages.** The example uses English. Tenants might need localized label strings and error messages.
   - **Recommendation:** ship v1 with single-locale strings; structure labels and messages as plain strings. Mark `label`, `helpText`, and `message` as localization-ready keys (`label.i18nKey` form) in v2 if requested.
   - **Tradeoff:** zero i18n complexity now vs rebuilding strings later.

10. **Audit log retention and access.** We proposed indefinite retention with no UI for non-admins. Should tenants be able to (i) configure retention (e.g., 7 years for compliance), (ii) export audit, (iii) view audit from inside the designer UI?
    - **Recommendation:** (iii) ships in v1 — audit panel in the designer, gated by `formbuilder.view`. (i) and (ii) are admin/ops concerns, defer to v2.
    - **Tradeoff:** in-product transparency vs operational tooling scope.

11. **`form_key` uniqueness strategy.** After dropping the `tenant_id` column, `UNIQUE (frm_key)` on `eamform_definitions` is enforced *within the tenant's schema only* (Section 2.0). Two tenants can each define a form called `customer-intake` and never collide. Within a single tenant's schema, should two forms be allowed to share the same `form_key`? The Section 2.0 model permits it (the unique constraint was already proposed as `UNIQUE (frm_key)`), but sanity argues against it: URLs (`/api/formbuilder/forms/:formKey`) would be ambiguous, code references would break, and the designer's form picker would show duplicates.
    - **Options:**
      - (a) **`UNIQUE (frm_key)`** within the schema — current proposal. Two forms cannot share a key. Strong default.
      - (b) **No uniqueness constraint**, allow duplicate keys — would force URL disambiguation (e.g. `?id=`) and is generally bad UX.
      - (c) **Composite uniqueness on `(frm_key, frm_status)`** — allow an archived form to be replaced by a new active one with the same key. Useful for "rename by archive + recreate" workflows.
    - **Recommendation:** (a). Single column, simple, no surprises. If (c) becomes useful later it is a one-line schema change (drop the unique, add the composite), and the application layer can be updated in the same PR.
    - **Tradeoff:** (a) blocks "rename by reuse" workflows; (c) supports them but adds complexity. Not a v1 concern.

---

## Appendix A — Terminology

| Term | Meaning |
|---|---|
| **Tenant** | In this module, a *logical* scope identified by the schema name (`tenant_<code>`). The schema is pinned by the `RunInTenantTx` middleware before any form-builder query runs; the module never sees a `tenant_id` value. See Section 2.0. |
| **Form definition** | The stable logical entity (`eamform_definitions`). Identified by `form_key` within the tenant's schema. Has no version pointer of its own — versions live on layouts. |
| **Layout** | A named, complete JSON artifact representing the UX for one role audience (`eamform_layouts`). Identified by `(form_id, name)` within the tenant's schema. |
| **`default` layout** | The reserved layout name within a form whose purpose is to be the implicit fallback for any role without an explicit assignment. Auto-created on form creation, cannot be renamed, cannot be hard-deleted while the form exists, and cannot be soft-archived while the form is `active`. Identified by `(form_id, name='default')` — like any other layout, but with service-layer invariants on top. |
| **Layout version** | A specific immutable snapshot of a layout (`eamform_layout_versions`). Identified by `version_number`. |
| **Draft** | The mutable working copy of a layout. At most one per layout. |
| **Published version** | The version of a layout currently served to end users. At most one per layout. |
| **Role assignment** | A row in `eamform_role_assignments` that says "for this form, this role is served by this layout". At most one active per `(form, role)`. The role is a key in the tenant's user-management system; it is not global across tenants. |
| **Resolved layout** | The full layout JSON returned by runtime — equal to the published version's `flv_definition`, no transformation. |
| **Audit entry** | A row in `eamform_audit_log` recording one mutation. Append-only, immutable. `actor_user_id` is the user within the tenant that owns the schema. |

## Appendix B — Note on the existing `query-builder`

`frontend/src/app/shared/components/query-builder/` is the visual builder for **grid queries** (which columns, which filters, which sorts) used inside grid screens. It is unrelated to data-entry forms. It happens to share Angular CDK DragDrop and signals — which is a *good* sign for the patterns we propose — but it solves a different problem and we do not consolidate it.
