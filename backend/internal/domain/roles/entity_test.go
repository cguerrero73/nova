package roles

import (
	"testing"
)

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions map[string]map[string]bool
		screen      string
		action      string
		want        bool
	}{
		{
			name:        "nil permissions returns false",
			permissions: nil,
			screen:      "formbuilder",
			action:      "design",
			want:        false,
		},
		{
			name:        "empty permissions returns false",
			permissions: map[string]map[string]bool{},
			screen:      "formbuilder",
			action:      "design",
			want:        false,
		},
		{
			name: "exact match returns true",
			permissions: map[string]map[string]bool{
				"formbuilder": {"design": true},
			},
			screen: "formbuilder",
			action: "design",
			want:   true,
		},
		{
			name: "exact match denied returns false",
			permissions: map[string]map[string]bool{
				"formbuilder": {"design": false},
			},
			screen: "formbuilder",
			action: "design",
			want:   false,
		},
		{
			name: "missing screen returns false",
			permissions: map[string]map[string]bool{
				"formbuilder": {"design": true},
			},
			screen: "parts",
			action: "design",
			want:   false,
		},
		{
			name: "missing action returns false",
			permissions: map[string]map[string]bool{
				"formbuilder": {"design": true},
			},
			screen: "formbuilder",
			action: "publish",
			want:   false,
		},
		{
			name: "wildcard screen grants all actions",
			permissions: map[string]map[string]bool{
				"*": {"*": true},
			},
			screen: "formbuilder",
			action: "design",
			want:   true,
		},
		{
			name: "wildcard screen with specific action",
			permissions: map[string]map[string]bool{
				"*": {"view": true},
			},
			screen: "formbuilder",
			action: "view",
			want:   true,
		},
		{
			name: "specific screen with wildcard action",
			permissions: map[string]map[string]bool{
				"formbuilder": {"*": true},
			},
			screen: "formbuilder",
			action: "anything",
			want:   true,
		},
		{
			name: "ADMIN wildcard overrides specific denial",
			permissions: map[string]map[string]bool{
				"*":           {"*": true},
				"formbuilder": {"design": false},
			},
			screen: "formbuilder",
			action: "design",
			want:   true, // wildcard * checked first
		},
		{
			name: "multiple screens with mixed permissions",
			permissions: map[string]map[string]bool{
				"formbuilder": {"view": true, "design": false},
				"parts":       {"view": true, "edit": true},
			},
			screen: "parts",
			action: "edit",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &Role{
				ID:          "TEST",
				Permissions: tt.permissions,
			}

			got := role.HasPermission(tt.screen, tt.action)
			if got != tt.want {
				t.Errorf("HasPermission(%q, %q) = %v, want %v", tt.screen, tt.action, got, tt.want)
			}
		})
	}
}

func TestIsActive(t *testing.T) {
	tests := []struct {
		name    string
		notUsed *string
		want    bool
	}{
		{
			name:    "nil NotUsed means active",
			notUsed: nil,
			want:    true,
		},
		{
			name:    "dash means active",
			notUsed: strPtr("-"),
			want:    true,
		},
		{
			name:    "plus means inactive",
			notUsed: strPtr("+"),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &Role{NotUsed: tt.notUsed}
			if got := role.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
