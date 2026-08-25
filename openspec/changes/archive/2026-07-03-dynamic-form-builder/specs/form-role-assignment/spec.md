# Form Role Assignment Specification

## Purpose

Map tenant roles to specific layouts per form. At most one active assignment per (form, role). Unassigned roles fall back to the `default` layout.

## Requirements

### Requirement: Role-to-Layout Assignment

The system MUST allow assigning a layout to a role for a given form. Only one active assignment per (form, role) pair is permitted.

#### Scenario: Assign layout to role

- GIVEN a form with layouts `default` and `admin-full`, and user has `formbuilder.assign`
- WHEN `PUT /forms/:formKey/assignments/admin` with `{ layoutName: "admin-full" }`
- THEN an active assignment row is created linking `admin` → `admin-full`
- AND an `layout.assign` audit entry is written

#### Scenario: Replace existing assignment

- GIVEN `admin` is already assigned to `admin-full`
- WHEN `PUT .../assignments/admin` with `{ layoutName: "viewer-readonly" }`
- THEN the previous assignment is revoked (`fra_revoked_at` set)
- AND a new active assignment is created for `viewer-readonly`

#### Scenario: Reject cross-form layout assignment

- GIVEN a layout that belongs to a different form
- WHEN a user attempts to assign it
- THEN the system MUST return `422 Unprocessable Entity`

### Requirement: Assignment Revocation

The system MUST support soft-revoking an assignment by setting `fra_revoked_at`. Revoked assignments MUST NOT affect runtime resolution.

#### Scenario: Revoke an assignment

- GIVEN an active assignment for `admin` on a form
- WHEN `DELETE /forms/:formKey/assignments/admin` with `formbuilder.assign`
- THEN `fra_revoked_at` is set to current timestamp
- AND a `layout.unassign` audit entry is written

#### Scenario: Revoke non-existent assignment

- GIVEN no active assignment for role `viewer` on a form
- WHEN `DELETE /forms/:formKey/assignments/viewer`
- THEN the system MUST return `404 Not Found`

### Requirement: Default Layout Fallback

When no active assignment exists for a role, the system MUST resolve to the layout named `default` for that form. The system MUST NOT return an error for missing assignments.

#### Scenario: Unassigned role falls back to default

- GIVEN a form with a published `default` layout and no assignment for role `viewer`
- WHEN runtime resolution is invoked for `viewer`
- THEN the `default` layout's published version is returned

#### Scenario: Explicit default assignment is valid

- GIVEN a designer explicitly assigns `default` layout to role `viewer`
- WHEN runtime resolution is invoked for `viewer`
- THEN the `default` layout is returned (same result as implicit fallback)
