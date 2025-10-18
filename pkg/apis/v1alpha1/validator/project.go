package validator

import (
	"context"
	"log/slog"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	auth0validator "github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/cockroachdb/errors"
	// v1alpha1generated "github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
)

func PreValidateProjectCreate(
	ctx context.Context,
	logger *slog.Logger) error {
	v, ok := ctx.Value(jwtmiddleware.ContextKey{}).(*auth0validator.ValidatedClaims)
	if !ok {
		return errors.New("the auth info is not valid")
	}
	logger.InfoContext(ctx, "", slog.Any("auth", v))

	return nil
}
