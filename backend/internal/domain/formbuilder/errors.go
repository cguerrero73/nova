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
