# Form Audit Specification

## Purpose

Append-only, immutable audit log for all form builder mutations. Records who did what, when, with action-specific metadata.

## Requirements

### Requirement: Automatic Audit on Mutations

The system MUST write an audit entry for every mutation: form create/archive, layout create/archive/assign/unassign, version draft_save/publish/revert.

#### Scenario: Form creation writes audit

- GIVEN a user creates a form
- THEN two audit entries are written: `form.create` and `layout.create` (for the auto-created `default` layout)

#### Scenario: Publish writes audit with metadata

- GIVEN a user publishes version 3 of a layout
- THEN a `version.publish` audit entry is written with `metadata: { version_number: 3, description: "..." }`

#### Scenario: Assignment change writes audit

- GIVEN a user assigns `admin-full` to role `admin`
- THEN a `layout.assign` audit entry is written with `metadata: { role: "admin", layout: "admin-full" }`

### Requirement: Audit Immutability

Audit entries MUST NOT be modifiable or deletable after insert. The system MUST enforce this at the database level via triggers.

#### Scenario: Attempt to update audit row

- GIVEN an audit entry exists
- WHEN an UPDATE is attempted on the row
- THEN the database trigger MUST raise an exception

#### Scenario: Attempt to delete audit row

- GIVEN an audit entry exists
- WHEN a DELETE is attempted on the row
- THEN the database trigger MUST raise an exception

### Requirement: Audit Retrieval

The system MUST provide a paged, filterable audit log endpoint. Filtering MUST support: `action`, `entity_type`, `actor_user_id`, and date range.

#### Scenario: List audit entries for a form

- GIVEN a form with 15 audit entries
- WHEN `GET /forms/:formKey/audit` with default pagination
- THEN the first page of entries is returned ordered by `created_at DESC`

#### Scenario: Filter by action type

- GIVEN audit entries with various action types
- WHEN `GET /forms/:formKey/audit?action=version.publish`
- THEN only `version.publish` entries are returned

### Requirement: Audit Actor Attribution

Every audit entry MUST record the `actor_user_id` from the authenticated user's session context.

#### Scenario: Actor recorded from session

- GIVEN user `jsmith` performs a layout creation
- WHEN the audit entry is written
- THEN `fal_actor_user_id = "jsmith"`
