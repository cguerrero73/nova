# Form Designer Specification

## Purpose

Visual designer for creating and editing form layouts. Provides drag-and-drop section/field authoring, field configuration, preview, and role-assignment management.

## Requirements

### Requirement: Visual Layout Editor

The system MUST provide a visual designer with three panels: field palette (left), section canvas (center), and field settings (right).

#### Scenario: Drag field from palette to section

- GIVEN the designer is open with a layout loaded
- WHEN a user drags a `text` field from the palette onto a section
- THEN the field appears in the section's field list
- AND the layout JSON is updated with the new field

#### Scenario: Reorder fields within a section

- GIVEN a section with 3 fields
- WHEN a user drags field 3 above field 1
- THEN the `order` values are updated to reflect the new sequence

#### Scenario: Move field between sections

- GIVEN two sections, each with fields
- WHEN a user drags a field from section A to section B
- THEN the field is removed from section A and added to section B with updated order

### Requirement: Field Type Catalog

The designer MUST support exactly 8 field types: `text`, `textarea`, `number`, `date`, `checkbox`, `select`, `radio`, `multiselect`.

#### Scenario: All field types available in palette

- GIVEN the designer is open
- WHEN the field palette is rendered
- THEN all 8 field types are visible and draggable

#### Scenario: Select field requires options

- GIVEN a `select` field is placed on the canvas
- WHEN the user attempts to save without defining options
- THEN validation MUST fail with an error indicating options are required

### Requirement: Shared Schema Validation

The designer MUST validate the layout JSON against the shared Zod schema before saving. The same schema MUST be used by the backend.

#### Scenario: Valid layout passes validation

- GIVEN a layout JSON conforming to the shared schema
- WHEN the user clicks "Save draft"
- THEN the layout is saved successfully

#### Scenario: Invalid layout rejected

- GIVEN a layout JSON missing required fields
- WHEN the user clicks "Save draft"
- THEN inline validation errors are displayed and the save is blocked

### Requirement: Layout and Assignment Management

The designer MUST display existing layouts for a form and allow creating new ones. The `default` layout MUST be visually distinguished. Assignment panel MUST show role-to-layout mappings.

#### Scenario: Default layout badge

- GIVEN a form with `default` and `admin-full` layouts
- WHEN the layout picker is rendered
- THEN `default` shows a "Default" badge with fallback description

#### Scenario: Assignment panel shows unassigned roles

- GIVEN roles `admin` (assigned to `admin-full`) and `viewer` (unassigned)
- WHEN the assignment panel is rendered
- THEN `admin` shows `admin-full` and `viewer` shows "→ uses `default` layout"

### Requirement: Preview Mode

The designer MUST allow previewing the current draft using the same renderer as runtime.

#### Scenario: Preview current draft

- GIVEN a layout with unsaved draft changes
- WHEN the user clicks "Preview"
- THEN a dialog renders the layout using `FormRuntimeComponent` with the in-memory draft
