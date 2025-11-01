package validator

import (
	"context"
	"log/slog"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	auth0validator "github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/cockroachdb/errors"
	v1alpha1generated "github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
	"github.com/tacokumo/admin-api/pkg/middleware"
)

func IsOnlyPermitedToReadPersonalProjects(ctx context.Context) (bool, error) {
	claims, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*auth0validator.ValidatedClaims)
	if !ok {
		return false, errors.New("the auth info is not valid")
	}
	customClaims, ok := claims.CustomClaims.(*middleware.CustomClaims)
	if !ok {
		return false, errors.New("invalid custom claims type")
	}
	perms, err := ParsePermissions(customClaims.Permissions)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if perms.PersonalProject.CanRead {
		if len(perms.Project) == 0 {
			return true, nil
		}
	}

	return false, nil
}

func CollectPermittedProjectNamesToRead(ctx context.Context) ([]string, error) {
	claims, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*auth0validator.ValidatedClaims)
	if !ok {
		return nil, errors.New("the auth info is not valid")
	}
	customClaims, ok := claims.CustomClaims.(*middleware.CustomClaims)
	if !ok {
		return nil, errors.New("invalid custom claims type")
	}
	perms, err := ParsePermissions(customClaims.Permissions)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	names := make([]string, 0, len(perms.Project))
	for name := range perms.Project {
		names = append(names, name)
	}
	return names, nil
}

func PreValidateProjectCreate(
	ctx context.Context,
	logger *slog.Logger,
	proj *v1alpha1generated.Project,
) error {
	claims, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*auth0validator.ValidatedClaims)
	if !ok {
		return errors.New("the auth info is not valid")
	}

	customClaims, ok := claims.CustomClaims.(*middleware.CustomClaims)
	if !ok {
		return errors.New("invalid custom claims type")
	}

	perms, err := ParsePermissions(customClaims.Permissions)
	if err != nil {
		return errors.WithStack(err)
	}

	switch proj.Kind {
	case v1alpha1generated.ProjectKindPersonal:
		if err := preValidatePersonalProjectCreate(ctx, logger, perms, proj); err != nil {
			return errors.WithStack(err)
		}
	case v1alpha1generated.ProjectKindShared:
		if err := preValidateSharedProjectCreate(ctx, logger, perms, proj); err != nil {
			return errors.WithStack(err)
		}
	default:
		return errors.Newf("invalid project kind: %s", proj.Kind)
	}

	return nil
}

func preValidatePersonalProjectCreate(
	_ context.Context,
	_ *slog.Logger,
	permissions Permissions,
	_ *v1alpha1generated.Project,
) error {
	if !permissions.PersonalProject.CanCreate {
		return errors.Newf("permission denied to create personal project")
	}
	return nil
}

func preValidateSharedProjectCreate(
	_ context.Context,
	_ *slog.Logger,
	permissions Permissions,
	proj *v1alpha1generated.Project,
) error {
	if !permissions.Project[proj.Name].CanCreate {
		return errors.Newf("permission denied to create shared project: %s", proj.Name)
	}

	return nil
}
