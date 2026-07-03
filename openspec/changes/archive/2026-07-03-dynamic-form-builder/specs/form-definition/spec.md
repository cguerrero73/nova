# Form Definition Specification

## Purpose

Manage the lifecycle of logical form definitions within a tenant schema. Each form auto-ships a reserved `default` layout on creation.

## Requirements

### Requirement: Form Creation with Default Layout

The system MUST create a form definition and, in the same transaction, auto-create a layout named `default` with `status=active` and no draft or published version.

#### Scenario: Successful form creation

- GIVEN an authenticated user with `formbuilder.design` permission
- WHEN `POST /api/formbuilder/forms` with `{ key, name, description }`
- THEN a form row is created in `eamform_definitions`
- AND a `default` layout row is created in `eamform_layouts` for that form
- AND a `layout.create` audit entry is written with `metadata.auto_created = true`

#### Scenario: Duplicate form key rejected

- GIVEN a form with `frm_key = "customer-intake"` already exists in the tenant schema
- WHEN a user attempts to create another form with the same key
- THEN the system MUST return `409 Conflict`

### Requirement: Form Key Uniqueness

The system MUST enforce `UNIQUE (frm_key)` within the tenant schema. No two active or archived forms MAY share the same key.

#### Scenario: Unique keys across forms

- GIVEN two forms with different keys
- THEN both exist independently in the tenant schema

### Requirement: Form Archival

The system MUST support soft-archiving a form. Archived forms MUST NOT be served at runtime and MUST NOT be editable.

#### Scenario: Archive an active form

- GIVEN an active form with `formbuilder.publish` permission
- WHEN `POST /api/formbuilder/forms/:formKey/archive`
- THEN `frm_status` becomes `archived`
- AND the form is excluded from runtime resolution

#### Scenario: Archived form not editable

- GIVEN an archived form
- WHEN a user attempts to edit its layouts
- THEN the system MUST return `410 Gone`
