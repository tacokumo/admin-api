package validator_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	auth0validator "github.com/auth0/go-jwt-middleware/v2/validator"
	v1alpha1generated "github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1/validator"
	"github.com/tacokumo/admin-api/pkg/middleware"
)

func TestPreValidateProjectCreate(t *testing.T) {
	tests := []struct {
		name        string
		testCtx     func() context.Context
		proj        *v1alpha1generated.Project
		expectedErr bool
	}{
		{
			name: "personal_projectの作成権限を持っているときはエラーにならないこと",
			testCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, &auth0validator.ValidatedClaims{
					CustomClaims: &middleware.CustomClaims{
						Permissions: []string{"personal_project:create"},
					},
				})
				return ctx
			},
			proj: &v1alpha1generated.Project{
				Kind: v1alpha1generated.ProjectKindPersonal,
			},
			expectedErr: false,
		},
		{
			name: "personal_projectの作成権限を持っていないときはエラーになること",
			testCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, &auth0validator.ValidatedClaims{
					CustomClaims: &middleware.CustomClaims{},
				})
				return ctx
			},
			proj: &v1alpha1generated.Project{
				Kind: v1alpha1generated.ProjectKindPersonal,
			},
			expectedErr: true,
		},
		{
			name: "shared_project fooの作成権限を持っているときはエラーにならないこと",
			testCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, &auth0validator.ValidatedClaims{
					CustomClaims: &middleware.CustomClaims{
						Permissions: []string{"project:foo:create"},
					},
				})
				return ctx
			},
			proj: &v1alpha1generated.Project{
				Kind: v1alpha1generated.ProjectKindShared,
				Name: "foo",
			},
			expectedErr: false,
		},
		{
			name: "shared_project fooの作成権限を持っていないときはエラーになること",
			testCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), jwtmiddleware.ContextKey{}, &auth0validator.ValidatedClaims{
					CustomClaims: &middleware.CustomClaims{
						Permissions: []string{"project:bar:create"},
					},
				})
				return ctx
			},
			proj: &v1alpha1generated.Project{
				Kind: v1alpha1generated.ProjectKindShared,
				Name: "foo",
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			err := validator.PreValidateProjectCreate(tt.testCtx(), logger, tt.proj)
			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}
