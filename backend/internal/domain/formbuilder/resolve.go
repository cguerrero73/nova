package formbuilder

import (
	"context"
	"fmt"
)

// Resolve performs runtime resolution: given a form key and role name, it
// returns the published layout version. The algorithm:
//  1. Look up the active assignment for the form + role.
//  2. If no assignment, fall back to the "default" layout.
//  3. Load the published version of the resolved layout.
//  4. If no published version exists, return ErrFormLayoutNotPublished.
//
// No tenantID parameter — RunInTenantTx already scoped the connection.
func (s *FormService) Resolve(ctx context.Context, formKey string, roleName string) (*ResolveResult, error) {
	// 1. Load form
	form, err := s.forms.FindByKey(ctx, formKey)
	if err != nil {
		return nil, fmt.Errorf("resolving form: %w", err)
	}
	if form == nil {
		return nil, ErrFormNotFound
	}
	if form.Status == "archived" {
		return nil, ErrFormArchived
	}

	// 2. Resolve layout: assignment → default fallback
	var layout *Layout

	if roleName != "" {
		assignment, err := s.assignments.FindActiveByFormAndRole(ctx, form.ID, roleName)
		if err != nil {
			return nil, fmt.Errorf("looking up assignment: %w", err)
		}
		if assignment != nil {
			layout, err = s.layouts.FindByID(ctx, assignment.LayoutID)
			if err != nil {
				return nil, fmt.Errorf("loading assigned layout: %w", err)
			}
			if layout == nil {
				return nil, ErrLayoutNotFound
			}
		}
	}

	// Fallback to default layout
	if layout == nil {
		layout, err = s.layouts.FindByFormAndName(ctx, form.ID, "default")
		if err != nil {
			return nil, fmt.Errorf("loading default layout: %w", err)
		}
		if layout == nil {
			return nil, ErrFormDefaultLayoutMissing
		}
	}

	// 3. Load published version
	if layout.PublishedVersionID == nil {
		return nil, ErrFormLayoutNotPublished
	}

	version, err := s.versions.FindByID(ctx, *layout.PublishedVersionID)
	if err != nil {
		return nil, fmt.Errorf("loading published version: %w", err)
	}
	if version == nil {
		return nil, ErrFormLayoutNotPublished
	}

	return &ResolveResult{
		FormKey:    formKey,
		LayoutName: layout.Name,
		Version:    version.VersionNumber,
		Definition: version.Definition,
	}, nil
}

// ResolveResult is the output of the Resolve algorithm.
type ResolveResult struct {
	FormKey    string `json:"formKey"`
	LayoutName string `json:"layoutName"`
	Version    int    `json:"version"`
	Definition []byte `json:"definition"`
}
