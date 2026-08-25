# Form Runtime Resolution Specification

## Purpose

Resolve the correct layout JSON for a given (form, role) pair at runtime. Uses assignment lookup with `default` fallback, returns the published version snapshot.

## Requirements

### Requirement: Runtime Resolution Algorithm

The system MUST resolve a form request by: (1) looking up the active assignment for the caller's role, (2) falling back to `default` if no assignment exists, (3) returning the published version's JSON.

#### Scenario: Assigned role receives its layout

- GIVEN form `customer-intake` with `admin` → `admin-full` assignment
- AND `admin-full` has published version 3
- WHEN `GET /api/formbuilder/forms/customer-intake` as user with active role `admin`
- THEN the response contains the complete JSON from published version 3 of `admin-full`

#### Scenario: Unassigned role receives default layout

- GIVEN form `customer-intake` with no assignment for role `viewer`
- AND `default` layout has published version 2
- WHEN `GET /api/formbuilder/forms/customer-intake` as user with active role `viewer`
- THEN the response contains the complete JSON from published version 2 of `default`

#### Scenario: Layout never published

- GIVEN a form where the resolved layout has no published version
- WHEN runtime resolution is invoked
- THEN the system MUST return `404 FormLayoutNotPublished`

#### Scenario: Default layout missing (data integrity)

- GIVEN a form where the `default` layout row does not exist (corrupt state)
- WHEN runtime resolution falls back to `default`
- THEN the system MUST return `500 FormDefaultLayoutMissing`

### Requirement: Resolution Caching

The system SHOULD cache resolved layouts with a TTL of 10 minutes. Cache keys MUST include the role name and version ID.

#### Scenario: Cache hit on repeated resolution

- GIVEN a layout was resolved and cached for role `admin`, version 3
- WHEN the same resolution is requested within 10 minutes
- THEN the cached JSON is returned without a database query

#### Scenario: Cache miss after publish

- GIVEN a cached entry for role `admin`, version 3
- WHEN version 4 is published (new version ID)
- THEN the next resolution for `admin` produces a cache miss and fetches version 4

### Requirement: Cross-Field Rules

The layout JSON MAY contain a `rules[]` array with cross-field validators. Supported operators: `equals`, `notEquals`, `requiredIf`, `hiddenIf`.

#### Scenario: equals rule validation

- GIVEN a rule `{ when: { path: "confirm_email" }, expect: { equals: { path: "email" } } }`
- WHEN the renderer evaluates the rule with mismatched values
- THEN the `confirm_email` field receives the rule's error message

#### Scenario: hiddenIf rule evaluation

- GIVEN a rule with `hiddenIf` operator
- WHEN the condition path has a truthy value
- THEN the target field is hidden from the rendered form

### Requirement: Tenant Isolation

All resolution queries MUST execute within `RunInTenantTx`. The system MUST NOT read or pass `tenant_id` values.

#### Scenario: Tenant-scoped resolution

- GIVEN two tenants each with a form keyed `customer-intake`
- WHEN tenant A requests `customer-intake`
- THEN only tenant A's schema is queried; tenant B's data is never accessed
