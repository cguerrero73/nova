package formbuilder

import "github.com/nova/backend/pkg/errors"

// Sentinel errors for form builder domain invariants.

// ErrReservedLayoutName is returned when a user attempts to create a layout named "default".
var ErrReservedLayoutName = errors.New("RESERVED_LAYOUT_NAME", "The layout name 'default' is reserved", 409)

// ErrCannotArchiveDefault is returned when attempting to archive the default layout of an active form.
var ErrCannotArchiveDefault = errors.New("CANNOT_ARCHIVE_DEFAULT", "Cannot archive the default layout while the form is active", 409)

// ErrFormDefaultLayoutMissing is returned when the default layout row is absent (data-integrity bug).
var ErrFormDefaultLayoutMissing = errors.New("FORM_DEFAULT_LAYOUT_MISSING", "Default layout is missing for this form (data integrity error)", 500)

// ErrFormLayoutNotPublished is returned when runtime resolution finds no published version.
var ErrFormLayoutNotPublished = errors.New("FORM_LAYOUT_NOT_PUBLISHED", "The resolved layout has no published version", 404)

// ErrFormArchived is returned when attempting to edit an archived form.
var ErrFormArchived = errors.New("FORM_ARCHIVED", "This form has been archived and cannot be modified", 410)

// ErrFormNotFound is returned when a form key does not match any definition.
var ErrFormNotFound = errors.New("FORM_NOT_FOUND", "Form not found", 404)

// ErrLayoutNotFound is returned when a layout name does not match any layout.
var ErrLayoutNotFound = errors.New("LAYOUT_NOT_FOUND", "Layout not found", 404)

// ErrFormKeyExists is returned when a duplicate form key is detected.
var ErrFormKeyExists = errors.New("FORM_KEY_EXISTS", "A form with this key already exists", 409)

// ErrLayoutNameExists is returned when a duplicate layout name is detected within a form.
var ErrLayoutNameExists = errors.New("LAYOUT_NAME_EXISTS", "A layout with this name already exists for this form", 409)

// ErrAssignmentNotFound is returned when no active assignment exists for the given role.
var ErrAssignmentNotFound = errors.New("ASSIGNMENT_NOT_FOUND", "No active assignment found for this role", 404)

// ErrNoDraft is returned when attempting to publish a layout that has no draft.
var ErrNoDraft = errors.New("NO_DRAFT", "No draft version exists for this layout", 422)

// ErrPublishDescriptionRequired is returned when publishing without a commit message.
var ErrPublishDescriptionRequired = errors.New("PUBLISH_DESCRIPTION_REQUIRED", "A description (commit message) is required to publish", 400)

// ErrVersionNotFound is returned when a specific version number does not exist.
var ErrVersionNotFound = errors.New("VERSION_NOT_FOUND", "The specified version does not exist", 404)

// ErrLayoutArchived is returned when attempting to modify an archived layout.
var ErrLayoutArchived = errors.New("LAYOUT_ARCHIVED", "This layout has been archived and cannot be modified", 410)

// ErrCrossFormLayout is returned when attempting to assign a layout from a different form.
var ErrCrossFormLayout = errors.New("CROSS_FORM_LAYOUT", "Cannot assign a layout that belongs to a different form", 422)
