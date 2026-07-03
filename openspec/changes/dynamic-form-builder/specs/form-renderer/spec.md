# Form Renderer Specification

## Purpose

Runtime renderer that takes a resolved layout JSON and produces an interactive form using Angular Reactive Forms. Renders sections, fields, and evaluates cross-field rules.

## Requirements

### Requirement: Layout Rendering

The system MUST render a complete layout JSON as an interactive form. Sections MUST be rendered in ascending `order`. Fields within sections MUST be rendered in ascending `order`.

#### Scenario: Render a multi-section layout

- GIVEN a layout with 3 sections (order 0, 1, 2) and varying fields
- WHEN the renderer receives the layout JSON
- THEN sections appear in order 0 → 1 → 2
- AND fields within each section appear in their defined order

#### Scenario: Read-only field rendering

- GIVEN a field with `ui.readOnly = true`
- WHEN the field is rendered
- THEN the field is displayed as non-editable

### Requirement: Field Type Rendering

The system MUST render each of the 8 field types using a dedicated renderer component. Each renderer MUST bind to Angular Reactive Forms controls.

#### Scenario: Text field renders input

- GIVEN a field with `type: "text"`
- WHEN rendered
- THEN an `<input type="text">` is displayed with label, placeholder, and helpText

#### Scenario: Select field renders dropdown

- GIVEN a field with `type: "select"` and static options
- WHEN rendered
- THEN a `<select>` dropdown is displayed with the defined choices

#### Scenario: Checkbox field renders toggle

- GIVEN a field with `type: "checkbox"`
- WHEN rendered
- THEN a checkbox input is displayed with the field label

### Requirement: Client-Side Validation

The system MUST translate layout `validators[]` into Angular Reactive Forms validators. Validation MUST run on value changes.

#### Scenario: Required field validation

- GIVEN a field with `{ kind: "required" }` validator
- WHEN the user leaves the field empty and blurs
- THEN a validation error is displayed

#### Scenario: Pattern validation

- GIVEN a field with `{ kind: "pattern", value: "^[A-Z]{3}$" }`
- WHEN the user enters `abc`
- THEN a validation error is displayed

### Requirement: Cross-Field Rule Evaluation

The system MUST evaluate `rules[]` from the layout JSON against the current FormGroup values on each value change.

#### Scenario: requiredIf rule

- GIVEN a rule `{ when: { path: "field_a" }, expect: { requiredIf: true }, target: "field_b" }`
- WHEN `field_a` has a truthy value
- THEN `field_b` becomes required and shows validation error if empty

#### Scenario: equals rule

- GIVEN a rule with `equals` operator comparing two fields
- WHEN the values differ
- THEN the target field displays the rule's error message

### Requirement: Grid Layout Support

The system MUST honor `ui.width` values (`full`, `half`, `third`) to arrange fields in a responsive grid.

#### Scenario: Half-width fields side by side

- GIVEN two consecutive fields with `ui.width: "half"`
- WHEN rendered
- THEN both fields appear side by side in a two-column layout
