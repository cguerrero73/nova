# Form Layout Specification

## Purpose

Manage named layouts per form. Each layout holds an independent version chain (draft → published → archived). Layouts are role-keyed, complete JSON artifacts.

## Requirements

### Requirement: Layout Creation

The system MUST allow creating named layouts within a form. The name `default` is reserved and MUST be rejected for user-created layouts.

#### Scenario: Create a new layout

- GIVEN a form exists and user has `formbuilder.design`
- WHEN `POST /forms/:formKey/layouts` with `{ name: "admin-full", displayName, description }`
- THEN a layout row is created with `status=active`, no draft, no published version
- AND a `layout.create` audit entry is written

#### Scenario: Reject reserved name `default`

- GIVEN a form already has the auto-created `default` layout
- WHEN a user attempts to create a layout named `default`
- THEN the system MUST return `409 ReservedLayoutName`

#### Scenario: Reject duplicate layout name

- GIVEN a layout `admin-full` exists for a form
- WHEN a user attempts to create another layout with the same name
- THEN the system MUST return `409 Conflict`

### Requirement: Draft Management

Each layout MUST support at most one draft version at a time. Saving a draft when none exists creates one; saving when one exists updates the draft.

#### Scenario: Save first draft

- GIVEN a layout with no draft
- WHEN `PUT /forms/:formKey/layouts/:layoutName/draft` with layout JSON
- THEN a new `draft` version row is created and `fl_draft_version_id` points to it

#### Scenario: Update existing draft

- GIVEN a layout with an existing draft
- WHEN `PUT .../draft` with updated JSON
- THEN the existing draft row's `flv_definition` is updated

### Requirement: Publish Immutable Snapshot

Publishing MUST create a new `published` version row from the current draft. Published versions MUST NOT be modified after insert.

#### Scenario: Publish a draft

- GIVEN a layout with a draft and user has `formbuilder.publish`
- WHEN `POST .../publish` with `{ description }`
- THEN a new `published` version row is created with monotonic `version_number`
- AND `fl_published_version_id` points to the new version
- AND the draft pointer is cleared
- AND a `version.publish` audit entry is written

#### Scenario: Published version immutability

- GIVEN a published version row exists
- WHEN any attempt is made to UPDATE or DELETE it
- THEN the database trigger MUST raise an exception

### Requirement: Layout Archival

The system MUST support soft-archiving a layout. The `default` layout MUST NOT be archived while its parent form is `active`.

#### Scenario: Archive a non-default layout

- GIVEN an active layout that is not `default`
- WHEN `POST .../archive` with `formbuilder.publish` permission
- THEN `fl_status` becomes `archived` and `fl_archived_at` is set

#### Scenario: Reject archive of default while form active

- GIVEN the `default` layout of an active form
- WHEN a user attempts to archive it
- THEN the system MUST return `409 CannotArchiveDefault`

### Requirement: Version History

The system MUST provide queryable version history per layout, including all published and archived versions.

#### Scenario: List version history

- GIVEN a layout with 3 published versions
- WHEN `GET .../versions`
- THEN the response contains all 3 versions ordered by `version_number DESC`

#### Scenario: Retrieve specific version

- GIVEN a published version with `version_number = 2`
- WHEN `GET .../versions/2`
- THEN the response contains the immutable JSON snapshot for that version
