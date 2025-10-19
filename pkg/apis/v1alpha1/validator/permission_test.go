package validator_test

import (
	"reflect"
	"testing"

	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1/validator"
)

func TestParsePermission(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		expected    validator.Permissions
	}{
		{
			name: "personal project create permission",
			permissions: []string{
				"personal_project:create",
			},
			expected: validator.Permissions{
				PersonalProject: validator.PersonalProjectPermissions{
					CanCreate: true,
				},
			},
		},
		{
			name: "project foo create permission",
			permissions: []string{
				"project:foo:create",
			},
			expected: validator.Permissions{
				Project: map[string]validator.ProjectPermissions{
					"foo": {
						CanCreate: true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms, err := validator.ParsePermissions(tt.permissions)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(tt.expected, perms) {
				t.Errorf("expected permissions %v, got %v", tt.expected, perms)
			}
		})
	}
}
